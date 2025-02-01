package main

import (
	"encoding/json"
	"fmt"
	"net"
)

func (s *StaticRoute) Validate() error {
	if s == nil {
		return fmt.Errorf("provided record is nil")
	}
	if s.Destination == "" {
		return fmt.Errorf("destination is required for static route")
	}
	if s.Gateway == "" {
		return fmt.Errorf("gateway is required for static route")
	}
	_, _, err := net.ParseCIDR(s.Destination)
	if err != nil {
		return fmt.Errorf("invalid address: %s", s.Destination)
	}
	gateway := net.ParseIP(s.Gateway)
	if gateway == nil {
		return fmt.Errorf("invalid gateway address: %s", s.Gateway)
	}
	return nil
}

func (s *StaticRoute) Create() error {
	if err := s.Validate(); err != nil {
		return err
	}
	// Check if the static route already exists
	exists, err := IsExistingStaticRoute(s.Destination)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("static route already exists")
	}
	tx := rwDB.Begin()
	defer tx.Rollback()
	err = tx.Create(s).Error
	if err != nil {
		return err
	}
	err = s.AddRoute()
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (s *StaticRoute) Delete() error {
	// If the static route does not exist, do nothing
	if exists, err := IsExistingStaticRoute(s.Destination); err != nil {
		return err
	} else if !exists {
		return nil
	}
	tx := rwDB.Begin()
	defer tx.Rollback()
	err := tx.Delete(&StaticRoute{}).Where("destination = ?", s.Destination).Error
	if err != nil {
		return err
	}
	err = s.RemoveRoute()
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func FetchStaticRouteByDestination(destination string) (*StaticRoute, error) {
	var route StaticRoute
	if err := rDB.Where("destination = ?", destination).First(&route).Error; err != nil {
		return nil, err
	}
	return &route, nil
}

func IsExistingStaticRoute(destination string) (bool, error) {
	var count int64
	if err := rDB.Model(&StaticRoute{}).Where("destination = ?", destination).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func FetchAllStaticRoutes() ([]StaticRoute, error) {
	var routes []StaticRoute
	if err := rDB.Find(&routes).Error; err != nil {
		return nil, err
	}
	return routes, nil
}

func (r *NFRule) Validate() error {
	if r == nil {
		return fmt.Errorf("provided record is nil")
	}
	if r.Table == "" {
		return fmt.Errorf("table is required for nf rule")
	}
	if r.Chain == "" {
		return fmt.Errorf("chain is required for nf rule")
	}
	if r.Args == "" {
		return fmt.Errorf("args is required for nf rule")
	}
	// Args should be a valid json string
	var args []string
	err := json.Unmarshal([]byte(r.Args), &args)
	if err != nil {
		return fmt.Errorf("invalid args: %v", err)
	}
	// args should not be empty
	if len(args) == 0 {
		return fmt.Errorf("args should not be empty")
	}
	return nil
}

func (r *NFRule) Create() error {
	if err := r.Validate(); err != nil {
		return err
	}
	tx := rwDB.Begin()
	defer tx.Rollback()
	err := tx.Create(r).Error
	if err != nil {
		return err
	}
	err = r.AddRule()
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (r *NFRule) Delete() error {
	if err := r.Validate(); err != nil {
		return err
	}
	tx := rwDB.Begin()
	defer tx.Rollback()
	err := tx.Delete(&NFRule{}).Where("uuid = ?", r.UUID).Error
	if err != nil {
		return err
	}
	err = r.RemoveRule()
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func FetchAllNFRules() ([]NFRule, error) {
	var rules []NFRule
	if err := rDB.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func FetchNFRuleByUUID(uuid string) (*NFRule, error) {
	var rule NFRule
	if err := rDB.Where("uuid = ?", uuid).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}
