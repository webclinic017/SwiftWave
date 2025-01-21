package main

import (
	"fmt"
)

func (d *DNSEntry) Validate() error {
	if d.Domain == "" || d.IP == "" {
		return fmt.Errorf("invalid dns entry")
	}
	return nil
}

func (d *DNSEntry) Create() error {
	if err := d.Validate(); err != nil {
		return err
	}
	if d.IP == "" {
		return fmt.Errorf("invalid dns entry")
	}
	exists, err := d.Exists()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return rwDB.Create(d).Error
}

func (d *DNSEntry) Delete() error {
	if err := d.Validate(); err != nil {
		return err
	}
	exists, err := d.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return rwDB.Delete(d).Error
}

func (d *DNSEntry) Exists() (bool, error) {
	if err := d.Validate(); err != nil {
		return false, err
	}
	var count int64
	err := rDB.Model(&DNSEntry{}).Where("domain = ?", d.Domain).Where("ip = ?", d.IP).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func FetchAllDNSRecords() ([]DNSEntry, error) {
	var dnsEntries []DNSEntry
	err := rDB.Find(&dnsEntries).Error
	return dnsEntries, err
}

func FetchDNSRecordsByDomain(domain string) ([]DNSEntry, error) {
	var dnsEntries []DNSEntry
	err := rDB.Where("domain = ?", domain).Find(&dnsEntries).Error
	return dnsEntries, err
}

func FetchARecordIps(domain string) []string {
	var ips []string
	records, err := FetchDNSRecordsByDomain(domain)
	if err != nil {
		return ips
	}
	for _, record := range records {
		ips = append(ips, record.IP)
	}
	return ips
}
