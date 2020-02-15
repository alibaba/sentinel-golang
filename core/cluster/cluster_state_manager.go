package cluster

import "time"

const(
	ClusterClient int32 = 0
	ClusterServer int32 = 1
	ClusterNotStarted int32 = -1
	MinINTERVAL int64 = 5 * 1000
)


type ClusterStateManager struct {
	Mode int32
	LastModified int
}


func NewClusterStateManager() *ClusterStateManager {
	return &ClusterStateManager{
		Mode:         ClusterNotStarted,
		LastModified: -1,
	}
}

func (csm *ClusterStateManager) IsClient() bool  {
	return csm.Mode == ClusterClient
}

func (csm * ClusterStateManager) IsServer() bool {
	return csm.Mode == ClusterServer
}

func (csm * ClusterStateManager) SetToClient () bool {
	if csm.Mode == ClusterClient{
		return true
	}
	csm.Mode = ClusterClient
	csm.SleepIfNeeded()
	csm.LastModified = time.Now().Nanosecond()
	return csm.StartClient()
}

func (csm *ClusterStateManager) SleepIfNeeded()  {
	if csm.LastModified <= 0{
		return
	}

	now := time.Now().Nanosecond()
	durationPast := now - csm.LastModified
	estimated := int64(durationPast) - MinINTERVAL

	if estimated < 0{
		time.Sleep(time.Nanosecond * time.Duration(estimated))
	}
}

func (csm *ClusterStateManager) StartClient() bool{
	//TODO
	
}
