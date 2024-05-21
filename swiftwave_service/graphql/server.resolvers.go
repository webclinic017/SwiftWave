package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.47

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/graphql/model"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	swiftwaveServiceManagerDocker "github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"gorm.io/gorm"
)

// CreateServer is the resolver for the createServer field.
func (r *mutationResolver) CreateServer(ctx context.Context, input model.NewServerInput) (*model.Server, error) {
	server := newServerInputToDatabaseObject(&input)
	err := core.CreateServer(&r.ServiceManager.DbClient, server)
	if err != nil {
		return nil, err
	}
	// if localhost, insert public key
	if server.IsLocalhost() {
		publicKey, err := r.Config.SystemConfig.PublicSSHKey()
		if err != nil {
			logger.GraphQLLoggerError.Println("Failed to generate public ssh key", err.Error())
		}
		// append the public key to current server ~/.ssh/authorized_keys
		err = AppendPublicSSHKeyLocally(publicKey)
		if err != nil {
			logger.GraphQLLoggerError.Println("Failed to append public ssh key", err.Error())
		}
	}
	return serverToGraphqlObject(server), nil
}

// DeleteServer is the resolver for the deleteServer field.
func (r *mutationResolver) DeleteServer(ctx context.Context, id uint) (bool, error) {
	// Checks -
	// 1. If server status `need_setup`, delete it from database
	// 2. If server status `preparing`, it can't be deleted
	// 3. If it's the last server, then delete it from db
	// 4. This should be swarm worker node
	// 5. Should disable ingress proxy on this server
	// 6. There should be another swarm manager running
	// 7. Remove from swarm cluster
	// 8. Remove from the database

	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// If `need_setup`, delete it from database
	if server.Status == core.ServerNeedsSetup {
		err = core.DeleteServer(&r.ServiceManager.DbClient, id)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	// If `preparing`, it can't be deleted
	if server.Status == core.ServerPreparing {
		return false, errors.New("server is preparing, you can delete it only after it come out of `preparing` status")
	}
	// If it's the last server, then delete it from db
	servers, err := core.FetchAllServers(&r.ServiceManager.DbClient)
	if err != nil {
		return false, err
	}
	if len(servers) == 1 {
		err = core.DeleteServer(&r.ServiceManager.DbClient, id)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	// if not the last server, then required additional steps
	if server.SwarmMode == core.SwarmManager {
		return false, errors.New("from 'Actions' menu, demote this server to 'Swarm Worker' mode to proceed for deletion")
	}
	if server.ProxyConfig.Enabled {
		return false, errors.New("from 'Actions' menu, disable ingress proxy on this server to proceed for deletion")
	}
	// fetch another swarm manager
	otherSwarmManager, err := core.FetchSwarmManagerExceptServer(&r.ServiceManager.DbClient, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("no other swarm manager found to proceed for deletion")
		}
	}
	// fetch docker manager
	dockerManager, err := swiftwaveServiceManagerDocker.DockerClient(ctx, otherSwarmManager)
	if err != nil {
		return false, err
	}
	// remove from swarm cluster
	err = dockerManager.RemoveNode(server.HostName)
	if err != nil {
		return false, err
	}
	// remove from local database
	err = core.DeleteServer(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// TestSSHAccessToServer is the resolver for the testSSHAccessToServer field.
func (r *mutationResolver) TestSSHAccessToServer(ctx context.Context, id uint) (bool, error) {
	command := "echo 'Hi'"
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	err = ssh_toolkit.ExecCommandOverSSH(command, nil, nil, 10, server.IP, server.SSHPort, server.User, r.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CheckDependenciesOnServer is the resolver for the checkDependenciesOnServer field.
func (r *mutationResolver) CheckDependenciesOnServer(ctx context.Context, id uint) ([]*model.Dependency, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Dependency, 0)
	for _, dependency := range core.RequiredServerDependencies {
		if dependency == "init" {
			continue
		}
		stdoutBuffer := new(bytes.Buffer)
		err = ssh_toolkit.ExecCommandOverSSH(core.DependencyCheckCommands[dependency], stdoutBuffer, nil, 5, server.IP, server.SSHPort, server.User, r.Config.SystemConfig.SshPrivateKey)
		if err != nil {
			if strings.Contains(err.Error(), "exited with status 1") {
				result = append(result, &model.Dependency{Name: dependency, Available: false})
				continue
			} else {
				return nil, err
			}
		}
		if stdoutBuffer.String() == "" {
			result = append(result, &model.Dependency{Name: dependency, Available: false})
		} else {
			result = append(result, &model.Dependency{Name: dependency, Available: true})
		}
	}
	return result, nil
}

// InstallDependenciesOnServer is the resolver for the installDependenciesOnServer field.
func (r *mutationResolver) InstallDependenciesOnServer(ctx context.Context, id uint) (bool, error) {
	_, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// Queue the request
	// - create a server log
	// - push the request to the queue
	serverLog := &core.ServerLog{
		ServerID: id,
		Title:    "Installing dependencies",
	}
	err = core.CreateServerLog(&r.ServiceManager.DbClient, serverLog)
	if err != nil {
		return false, err
	}
	// Push the request to the queue
	err = r.WorkerManager.EnqueueInstallDependenciesOnServerRequest(id, serverLog.ID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetupServer is the resolver for the setupServer field.
func (r *mutationResolver) SetupServer(ctx context.Context, input model.ServerSetupInput) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, input.ID)
	if err != nil {
		return false, err
	}

	// Check if advertise IP is valid ipv4
	if net.ParseIP(input.AdvertiseIP) == nil || net.ParseIP(input.AdvertiseIP).To4() == nil {
		return false, errors.New("invalid advertise IP")
	}

	// Check if all dependencies are installed
	installedDependencies, err := r.CheckDependenciesOnServer(ctx, input.ID)
	if err != nil {
		return false, err
	}
	for _, dependency := range installedDependencies {
		if !dependency.Available {
			return false, errors.New("dependency " + dependency.Name + " is not installed")
		}
	}

	// Proceed request logic (reject in any other case)
	// - if, want to be manager
	//    - if, there are some managers already, need to be online any of them
	//    - if, no servers, then it will be the first manager
	// - if, want to be worker
	//   - there need to be at least one manager
	var swarmManagerServer *core.Server
	if input.SwarmMode == model.SwarmModeManager {
		// Check if there are some servers already
		exists, err := core.IsPreparedServerExists(&r.ServiceManager.DbClient)
		if err != nil {
			return false, err
		}
		if exists {
			// Try to find out if there is any manager online
			r, err := core.FetchSwarmManager(&r.ServiceManager.DbClient)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return false, errors.New("swarm manager not found")
				} else {
					return false, err
				}
			}
			swarmManagerServer = &r
		}
	} else {
		// Check if there is any manager
		r, err := core.FetchSwarmManager(&r.ServiceManager.DbClient)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, errors.New("can't setup as worker, no swarm manager found")
			}
			return false, err
		}
		swarmManagerServer = &r
	}

	if swarmManagerServer == nil && input.SwarmMode == model.SwarmModeWorker {
		return false, errors.New("no manager found")
	}

	// NOTE: From here, if `swarmManagerServer` is nil, then this new server can be initialized as first swarm manager

	// Fetch hostname
	hostnameStdoutBuffer := new(bytes.Buffer)
	err = ssh_toolkit.ExecCommandOverSSH("cat /etc/hostname", hostnameStdoutBuffer, nil, 10, server.IP, server.SSHPort, server.User, r.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		return false, err
	}
	hostname := strings.TrimSpace(hostnameStdoutBuffer.String())
	server.HostName = hostname
	server.Status = core.ServerPreparing
	server.SwarmMode = core.SwarmMode(input.SwarmMode)
	server.DockerUnixSocketPath = input.DockerUnixSocketPath
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	if err != nil {
		return false, err
	}

	// Enqueue the request
	// - create a server log
	// - push the request to the queue
	serverLog := &core.ServerLog{
		ServerID: input.ID,
		Title:    "Setup server",
	}
	err = core.CreateServerLog(&r.ServiceManager.DbClient, serverLog)
	if err != nil {
		return false, err
	}
	// Push the request to the queue
	err = r.WorkerManager.EnqueueSetupServerRequest(input.ID, serverLog.ID, input.AdvertiseIP)
	if err != nil {
		return false, err
	}
	return true, nil
}

// PromoteServerToManager is the resolver for the promoteServerToManager field.
func (r *mutationResolver) PromoteServerToManager(ctx context.Context, id uint) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	if server.Status != core.ServerOnline {
		return false, errors.New("server is not online")
	}
	// Fetch any swarm manager
	swarmManagerServer, err := core.FetchSwarmManagerExceptServer(&r.ServiceManager.DbClient, server.ID)
	if err != nil {
		return false, errors.New("no manager found")
	}
	// If there is any swarm manager, then promote this server to manager
	// Fetch net.Conn to the swarm manager
	conn, err := ssh_toolkit.NetConnOverSSH("unix", swarmManagerServer.DockerUnixSocketPath, 5, swarmManagerServer.IP, swarmManagerServer.SSHPort, swarmManagerServer.User, r.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		return false, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.GraphQLLoggerError.Println(err.Error())
		}
	}(conn)
	// Promote this server to manager
	manager, err := containermanger.New(ctx, conn)
	if err != nil {
		return false, err
	}
	err = manager.PromoteToManager(server.HostName)
	if err != nil {
		return false, err
	}
	server.SwarmMode = core.SwarmManager
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	return err == nil, err
}

// DemoteServerToWorker is the resolver for the demoteServerToWorker field.
func (r *mutationResolver) DemoteServerToWorker(ctx context.Context, id uint) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	if server.Status != core.ServerOnline {
		return false, errors.New("server is not online")
	}
	// Fetch any swarm manager
	swarmManagerServer, err := core.FetchSwarmManagerExceptServer(&r.ServiceManager.DbClient, server.ID)
	if err != nil {
		return false, errors.New("no manager found")
	}
	// If there is any swarm manager, then promote this server to manager
	// Fetch net.Conn to the swarm manager
	conn, err := ssh_toolkit.NetConnOverSSH("unix", swarmManagerServer.DockerUnixSocketPath, 5, swarmManagerServer.IP, swarmManagerServer.SSHPort, swarmManagerServer.User, r.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		return false, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.GraphQLLoggerError.Println(err.Error())
		}
	}(conn)
	// Promote this server to manager
	manager, err := containermanger.New(ctx, conn)
	if err != nil {
		return false, err
	}
	err = manager.DemoteToWorker(server.HostName)
	if err != nil {
		return false, err
	}
	server.SwarmMode = core.SwarmWorker
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	return err == nil, err
}

// RestrictDeploymentOnServer is the resolver for the restrictDeploymentOnServer field.
func (r *mutationResolver) RestrictDeploymentOnServer(ctx context.Context, id uint) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	oldScheduleDeploymentsStatus := server.ScheduleDeployments
	server.ScheduleDeployments = false
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	if err != nil {
		return false, err
	}
	// Enqueue request to update applications on schedule_deployment update
	err = r.WorkerManager.EnqueueUpdateApplicationOnServerScheduleDeploymentUpdateRequest(server.ID)
	if err != nil {
		// rollback status
		server.ScheduleDeployments = oldScheduleDeploymentsStatus
		_ = core.UpdateServer(&r.ServiceManager.DbClient, server)
		return false, err
	}
	return true, nil
}

// AllowDeploymentOnServer is the resolver for the allowDeploymentOnServer field.
func (r *mutationResolver) AllowDeploymentOnServer(ctx context.Context, id uint) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	oldScheduleDeploymentsStatus := server.ScheduleDeployments
	server.ScheduleDeployments = true
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	if err != nil {
		return false, err
	}
	// Enqueue request to update applications on schedule_deployment update
	err = r.WorkerManager.EnqueueUpdateApplicationOnServerScheduleDeploymentUpdateRequest(server.ID)
	if err != nil {
		// rollback status
		server.ScheduleDeployments = oldScheduleDeploymentsStatus
		_ = core.UpdateServer(&r.ServiceManager.DbClient, server)
		return false, err
	}
	return true, nil
}

// PutServerInMaintenanceMode is the resolver for the putServerInMaintenanceMode field.
func (r *mutationResolver) PutServerInMaintenanceMode(ctx context.Context, id uint) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	if server.Status != core.ServerOnline {
		return false, errors.New("server is not online")
	}
	// Fetch any swarm manager
	swarmManagerServer, err := core.FetchSwarmManager(&r.ServiceManager.DbClient)
	if err != nil {
		return false, errors.New("no manager found")
	}
	// If there is any swarm manager, then promote this server to manager
	// Fetch net.Conn to the swarm manager
	conn, err := ssh_toolkit.NetConnOverSSH("unix", swarmManagerServer.DockerUnixSocketPath, 5, swarmManagerServer.IP, swarmManagerServer.SSHPort, swarmManagerServer.User, r.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		return false, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.GraphQLLoggerError.Println(err.Error())
		}
	}(conn)
	manager, err := containermanger.New(ctx, conn)
	if err != nil {
		return false, err
	}
	err = manager.MarkNodeAsDrained(server.HostName)
	if err != nil {
		return false, err
	}
	server.MaintenanceMode = true
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	return err == nil, err
}

// PutServerOutOfMaintenanceMode is the resolver for the putServerOutOfMaintenanceMode field.
func (r *mutationResolver) PutServerOutOfMaintenanceMode(ctx context.Context, id uint) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	if server.Status != core.ServerOnline {
		return false, errors.New("server is not online")
	}
	// Fetch any swarm manager
	swarmManagerServer, err := core.FetchSwarmManager(&r.ServiceManager.DbClient)
	if err != nil {
		return false, errors.New("no manager found")
	}
	// If there is any swarm manager, then promote this server to manager
	// Fetch net.Conn to the swarm manager
	conn, err := ssh_toolkit.NetConnOverSSH("unix", swarmManagerServer.DockerUnixSocketPath, 5, swarmManagerServer.IP, swarmManagerServer.SSHPort, swarmManagerServer.User, r.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		return false, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.GraphQLLoggerError.Println(err.Error())
		}
	}(conn)
	manager, err := containermanger.New(ctx, conn)
	if err != nil {
		return false, err
	}
	err = manager.MarkNodeAsActive(server.HostName)
	if err != nil {
		return false, err
	}
	server.MaintenanceMode = false
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	return err == nil, err
}

// RemoveServerFromSwarmCluster is the resolver for the removeServerFromSwarmCluster field.
func (r *mutationResolver) RemoveServerFromSwarmCluster(ctx context.Context, id uint) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// Fetch any swarm manager
	swarmManagerServer, err := core.FetchSwarmManagerExceptServer(&r.ServiceManager.DbClient, server.ID)
	if err != nil {
		return false, errors.New("no manager found")
	}
	// If there is any swarm manager, then promote this server to manager
	// Fetch net.Conn to the swarm manager
	conn, err := ssh_toolkit.NetConnOverSSH("unix", swarmManagerServer.DockerUnixSocketPath, 5, swarmManagerServer.IP, swarmManagerServer.SSHPort, swarmManagerServer.User, r.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		return false, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.GraphQLLoggerError.Println(err.Error())
		}
	}(conn)
	manager, err := containermanger.New(ctx, conn)
	if err != nil {
		return false, err
	}
	err = manager.RemoveNode(server.HostName)
	if err != nil {
		return false, err
	}
	server.Status = core.ServerNeedsSetup
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	if err == nil {
		// try to connect to the server and leave from the swarm
		serverConn, err2 := ssh_toolkit.NetConnOverSSH("unix", server.DockerUnixSocketPath, 5, server.IP, swarmManagerServer.SSHPort, server.User, r.Config.SystemConfig.SshPrivateKey)
		if err2 == nil {
			defer func(serverConn net.Conn) {
				_ = serverConn.Close()
			}(serverConn)
			serverDockerManager, err2 := containermanger.New(ctx, serverConn)
			if err2 == nil {
				_ = serverDockerManager.LeaveSwarm()
			}
		}

	}
	return err == nil, err
}

// EnableProxyOnServer is the resolver for the enableProxyOnServer field.
func (r *mutationResolver) EnableProxyOnServer(ctx context.Context, id uint, typeArg model.ProxyType) (bool, error) {
	// Fetch the server
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	if server.ProxyConfig.Enabled {
		return false, errors.New("proxy is already enabled")
	}
	// Set the proxy type
	server.ProxyConfig.Type = core.ProxyType(typeArg)
	// update in db
	err = core.ChangeProxyType(&r.ServiceManager.DbClient, server, server.ProxyConfig.Type)
	if err != nil {
		return false, err
	}
	// Enable the proxy
	server.ProxyConfig.SetupRunning = true
	// For backup proxy, atleast 1 active proxy is required
	if server.ProxyConfig.Type == core.BackupProxy {
		activeProxies, err := core.FetchProxyActiveServers(&r.ServiceManager.DbClient)
		if err != nil {
			return false, err
		}
		if len(activeProxies) == 0 {
			return false, errors.New("for adding backup proxy, atleast 1 active proxy is required")
		}
	}
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	if err != nil {
		return false, err
	}
	// Create a server log
	serverLog := &core.ServerLog{
		ServerID: id,
		Title:    "Enable proxy on server " + server.HostName,
	}
	err = core.CreateServerLog(&r.ServiceManager.DbClient, serverLog)
	if err != nil {
		return false, err
	}
	// Queue the request
	err = r.WorkerManager.EnqueueSetupAndEnableProxyRequest(id, serverLog.ID)
	return err == nil, err
}

// DisableProxyOnServer is the resolver for the disableProxyOnServer field.
func (r *mutationResolver) DisableProxyOnServer(ctx context.Context, id uint) (bool, error) {
	// Fetch the server
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	// Disable the proxy
	server.ProxyConfig.Enabled = false
	server.ProxyConfig.SetupRunning = false
	err = core.UpdateServer(&r.ServiceManager.DbClient, server)
	return err == nil, err
}

// FetchAnalyticsServiceToken is the resolver for the fetchAnalyticsServiceToken field.
func (r *mutationResolver) FetchAnalyticsServiceToken(ctx context.Context, id uint, rotate bool) (string, error) {
	var tokenRecord *core.AnalyticsServiceToken
	var err error
	if !rotate {
		tokenRecord, err = core.FetchAnalyticsServiceToken(ctx, r.ServiceManager.DbClient, id)
	} else {
		tokenRecord, err = core.RotateAnalyticsServiceToken(ctx, r.ServiceManager.DbClient, id)
	}
	if err != nil {
		return "", err
	} else {
		return tokenRecord.IDToken()
	}
}

// ChangeServerIPAddress is the resolver for the changeServerIpAddress field.
func (r *mutationResolver) ChangeServerIPAddress(ctx context.Context, id uint, ip string) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	ip = strings.TrimSpace(ip)
	if len(ip) == 0 {
		return false, errors.New("IP is required")
	}
	if strings.Compare(server.IP, ip) == 0 {
		return false, errors.New("IP is already " + server.IP)
	}
	err = core.ChangeServerIP(&r.ServiceManager.DbClient, server, ip)
	if err != nil {
		return false, err
	}
	// Exit process
	logger.GraphQLLoggerError.Println("Server " + server.HostName + " IP changed to " + ip + "\nRestarting swiftwave in 2 seconds to take effect")
	// Restart swiftwave service
	go func() {
		<-time.After(2 * time.Second)
		color.Green("Restarting swiftwave service")
		color.Yellow("Swiftwave service will be restarted in 2 seconds")
		color.Yellow("If you are running without enabling service, run `swiftwave start` to start the service")
		_ = exec.Command("systemctl", "restart", "swiftwave.service").Run()
		os.Exit(0)
	}()
	return true, nil
}

// ChangeServerSSHPort is the resolver for the changeServerSSHPort field.
func (r *mutationResolver) ChangeServerSSHPort(ctx context.Context, id uint, port int) (bool, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return false, err
	}
	if server.SSHPort == port {
		return false, errors.New("SSH port is already " + fmt.Sprintf("%d", port))
	}
	err = core.ChangeSSHPort(&r.ServiceManager.DbClient, server, port)
	if err != nil {
		return false, err
	}
	// Exit process
	logger.GraphQLLoggerError.Println("Server " + server.HostName + " SSH port changed to " + fmt.Sprintf("%d", port) + "\nRestarting swiftwave in 2 seconds to take effect")
	// Restart swiftwave service
	go func() {
		<-time.After(2 * time.Second)
		color.Green("Restarting swiftwave service")
		color.Yellow("Swiftwave service will be restarted in 2 seconds")
		color.Yellow("If you are running without enabling service, run `swiftwave start` to start the service")
		_ = exec.Command("systemctl", "restart", "swiftwave.service").Run()
		os.Exit(0)
	}()
	return true, nil
}

// NoOfServers is the resolver for the noOfServers field.
func (r *queryResolver) NoOfServers(ctx context.Context) (int, error) {
	return core.NoOfServers(&r.ServiceManager.DbClient)
}

// NoOfPreparedServers is the resolver for the noOfPreparedServers field.
func (r *queryResolver) NoOfPreparedServers(ctx context.Context) (int, error) {
	return core.NoOfPreparedServers(&r.ServiceManager.DbClient)
}

// Servers is the resolver for the servers field.
func (r *queryResolver) Servers(ctx context.Context) ([]*model.Server, error) {
	servers, err := core.FetchAllServers(&r.ServiceManager.DbClient)
	if err != nil {
		return nil, err
	}
	serverList := make([]*model.Server, 0)
	for _, server := range servers {
		serverList = append(serverList, serverToGraphqlObject(&server))
	}
	return serverList, nil
}

// Server is the resolver for the server field.
func (r *queryResolver) Server(ctx context.Context, id uint) (*model.Server, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	return serverToGraphqlObject(server), nil
}

// PublicSSHKey is the resolver for the publicSSHKey field.
func (r *queryResolver) PublicSSHKey(ctx context.Context) (string, error) {
	return r.Config.SystemConfig.PublicSSHKey()
}

// ServerResourceAnalytics is the resolver for the serverResourceAnalytics field.
func (r *queryResolver) ServerResourceAnalytics(ctx context.Context, id uint, timeframe model.ServerResourceAnalyticsTimeframe) ([]*model.ServerResourceAnalytics, error) {
	var previousTime = time.Now()
	switch timeframe {
	case model.ServerResourceAnalyticsTimeframeLast1Hour:
		previousTime = time.Now().Add(-1 * time.Hour)
	case model.ServerResourceAnalyticsTimeframeLast3Hours:
		previousTime = time.Now().Add(-3 * time.Hour)
	case model.ServerResourceAnalyticsTimeframeLast6Hours:
		previousTime = time.Now().Add(-6 * time.Hour)
	case model.ServerResourceAnalyticsTimeframeLast12Hours:
		previousTime = time.Now().Add(-12 * time.Hour)
	case model.ServerResourceAnalyticsTimeframeLast24Hours:
		previousTime = time.Now().Add(-24 * time.Hour)
	case model.ServerResourceAnalyticsTimeframeLast7Days:
		previousTime = time.Now().Add(-7 * 24 * time.Hour)
	case model.ServerResourceAnalyticsTimeframeLast30Days:
		previousTime = time.Now().Add(-30 * 24 * time.Hour)
	}
	previousTimeUnix := previousTime.Unix()

	// fetch the server resource analytics
	serverResourceStat, err := core.FetchServerResourceAnalytics(ctx, r.ServiceManager.DbClient, id, uint(previousTimeUnix))
	if err != nil {
		return nil, err
	}
	// convert the server resource analytics to graphql object
	serverResourceStatList := make([]*model.ServerResourceAnalytics, 0)
	for _, record := range serverResourceStat {
		serverResourceStatList = append(serverResourceStatList, serverResourceStatToGraphqlObject(record))
	}
	return serverResourceStatList, nil
}

// ServerDiskUsage is the resolver for the serverDiskUsage field.
func (r *queryResolver) ServerDiskUsage(ctx context.Context, id uint) ([]*model.ServerDisksUsage, error) {
	// fetch the server disk usage
	serverResourceUsageRecords, err := core.FetchServerDiskUsage(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	serverDiskStatsList := make([]*model.ServerDisksUsage, 0)
	for _, record := range serverResourceUsageRecords {
		val := severDisksStatToGraphqlObject(record.DiskStats, record.RecordedAt)
		serverDiskStatsList = append(serverDiskStatsList, &val)
	}
	return serverDiskStatsList, nil
}

// ServerLatestResourceAnalytics is the resolver for the serverLatestResourceAnalytics field.
func (r *queryResolver) ServerLatestResourceAnalytics(ctx context.Context, id uint) (*model.ServerResourceAnalytics, error) {
	// fetch the latest server resource analytics
	serverResourceStat, err := core.FetchLatestServerResourceAnalytics(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	// convert the server resource analytics to graphql object
	return serverResourceStatToGraphqlObject(serverResourceStat), nil
}

// ServerLatestDiskUsage is the resolver for the serverLatestDiskUsage field.
func (r *queryResolver) ServerLatestDiskUsage(ctx context.Context, id uint) (*model.ServerDisksUsage, error) {
	// fetch the latest server disk usage
	serverDiskStats, timestamp, err := core.FetchLatestServerDiskUsage(ctx, r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	// convert the server disk usage to graphql object
	res := severDisksStatToGraphqlObject(*serverDiskStats, *timestamp)
	return &res, nil
}

// NetworkInterfacesOnServer is the resolver for the networkInterfacesOnServer field.
func (r *queryResolver) NetworkInterfacesOnServer(ctx context.Context, id uint) ([]*model.NetworkInterface, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, id)
	if err != nil {
		return nil, err
	}
	stdoutBuffer := new(bytes.Buffer)
	err = ssh_toolkit.ExecCommandOverSSH("ip -o addr show | awk '{print $2, $4}'", stdoutBuffer, nil, 5, server.IP, server.SSHPort, server.User, r.Config.SystemConfig.SshPrivateKey)
	if err != nil {
		return nil, err
	}
	rawLines := strings.Split(stdoutBuffer.String(), "\n")
	networkInterfaces := make([]*model.NetworkInterface, 0)
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		splits := strings.Split(line, " ")
		if len(splits) != 2 {
			continue
		}
		interfaceName := splits[0]
		// ignore `docker0` and `docker_gwbridge`
		if interfaceName == "docker0" || interfaceName == "docker_gwbridge" || interfaceName == "lo" {
			continue
		}
		ip := splits[1]
		// strip the cidr from the ip
		ip = strings.Split(ip, "/")[0]
		// check if ipv4
		if net.ParseIP(ip) != nil && net.ParseIP(ip).To4() != nil {
			networkInterfaces = append(networkInterfaces, &model.NetworkInterface{
				Name: interfaceName,
				IP:   ip,
			})
		}
	}
	return networkInterfaces, nil
}

// SwarmNodeStatus is the resolver for the swarmNodeStatus field.
func (r *serverResolver) SwarmNodeStatus(ctx context.Context, obj *model.Server) (string, error) {
	server, err := core.FetchServerByID(&r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return "", nil
	}
	if server.Status != core.ServerOnline {
		return "", nil
	}
	// Fetch any swarm manager
	swarmManagerServer, err := core.FetchSwarmManager(&r.ServiceManager.DbClient)
	if err != nil {
		return "", nil
	}
	manager, err := swiftwaveServiceManagerDocker.DockerClient(ctx, swarmManagerServer)
	if err != nil {
		return "", nil
	}
	return manager.FetchNodeStatus(server.HostName)
}

// Logs is the resolver for the logs field.
func (r *serverResolver) Logs(ctx context.Context, obj *model.Server) ([]*model.ServerLog, error) {
	serverLogs, err := core.FetchServerLogByServerID(&r.ServiceManager.DbClient, obj.ID)
	if err != nil {
		return nil, err
	}
	serverLogList := make([]*model.ServerLog, 0)
	for _, serverLog := range serverLogs {
		serverLogList = append(serverLogList, serverLogToGraphqlObject(&serverLog))
	}
	return serverLogList, nil
}

// Server returns ServerResolver implementation.
func (r *Resolver) Server() ServerResolver { return &serverResolver{r} }

type serverResolver struct{ *Resolver }
