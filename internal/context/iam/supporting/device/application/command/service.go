package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// Service manages device fingerprints.
type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// Register registers or updates a device fingerprint.
func (s *Service) Register(ctx context.Context, userID, deviceID, deviceName, ipAddress string) (_ *domain.Fingerprint, err error) {
	defer metrics.ObserveService("DeviceService", "Register", time.Now(), &err)

	fp := &domain.Fingerprint{
		UserID:     userID,
		DeviceHash: domain.HashDevice(deviceID),
		DeviceName: deviceName,
		IPAddress:  ipAddress,
		Trusted:    false,
	}

	if err = s.repo.Upsert(ctx, fp); err != nil {
		return nil, err
	}
	return fp, nil
}

// Trust marks a device as trusted.
func (s *Service) Trust(ctx context.Context, userID, deviceHash string) (err error) {
	defer metrics.ObserveService("DeviceService", "Trust", time.Now(), &err)

	fp, err := s.repo.GetByUserAndDevice(ctx, userID, deviceHash)
	if err != nil {
		return err
	}
	if fp == nil {
		return nil
	}
	fp.Trusted = true
	return s.repo.Upsert(ctx, fp)
}

// Untrust marks a device as untrusted.
func (s *Service) Untrust(ctx context.Context, userID, deviceHash string) (err error) {
	defer metrics.ObserveService("DeviceService", "Untrust", time.Now(), &err)

	fp, err := s.repo.GetByUserAndDevice(ctx, userID, deviceHash)
	if err != nil {
		return err
	}
	if fp == nil {
		return nil
	}
	fp.Trusted = false
	return s.repo.Upsert(ctx, fp)
}

// List returns all devices for a user.
func (s *Service) List(ctx context.Context, userID string) (_ []*domain.Fingerprint, err error) {
	defer metrics.ObserveService("DeviceService", "List", time.Now(), &err)
	return s.repo.ListByUserID(ctx, userID)
}

// Remove deletes a device fingerprint.
func (s *Service) Remove(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("DeviceService", "Remove", time.Now(), &err)
	return s.repo.Delete(ctx, id)
}
