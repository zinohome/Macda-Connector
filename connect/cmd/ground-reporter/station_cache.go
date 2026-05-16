package main

import "sync"

// StationInfo holds the latest station data for a device, extracted from NB67 protocol fields.
type StationInfo struct {
	CurStation  uint16 // cur_station from binary protocol (current station ID)
	NextStation uint16 // next_station from binary protocol (next station ID)
}

// StationCache stores the most recent StationInfo per device, updated from signal-parsed messages.
// Lookup at alarm/life event time is always "best-effort": returns zero values when the cache
// hasn't seen a parsed frame for this device yet (edge case at cold start).
type StationCache struct {
	mu    sync.RWMutex
	cache map[string]StationInfo // deviceID → latest station info
}

func newStationCache() *StationCache {
	return &StationCache{cache: make(map[string]StationInfo)}
}

// Update refreshes the station info for a device from a signal-parsed message's raw fields.
// Keys "CurStation" and "NextStation" match the PascalCase field names of the Kaitai-generated Nb67 struct.
func (c *StationCache) Update(msg ParsedMsg) {
	cur := uint16(rawInt(msg.Raw, "CurStation"))
	next := uint16(rawInt(msg.Raw, "NextStation"))
	if cur == 0 && next == 0 {
		return // device hasn't populated station fields yet
	}
	c.mu.Lock()
	c.cache[msg.DeviceID] = StationInfo{CurStation: cur, NextStation: next}
	c.mu.Unlock()
}

// Get returns the latest StationInfo for the given device, or a zero-value struct if unknown.
func (c *StationCache) Get(deviceID string) StationInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[deviceID]
}
