package service

import "time"

type AuditEntry struct {
	Action    string
	ActorID   string
	Resource  string
	Timestamp time.Time
}

type AuditService struct{}

func NewAuditService() *AuditService { return &AuditService{} }

func (s *AuditService) Record(action, actorID, resource string) AuditEntry {
	return AuditEntry{Action: action, ActorID: actorID, Resource: resource, Timestamp: time.Now()}
}
