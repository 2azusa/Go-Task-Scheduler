package etcdclient

const (
	keyEtcdProfile = "/pulse/"

	//key /pulse/node/<node_uuid>
	KeyEtcdNodeProfile = keyEtcdProfile + "node/"
	KeyEtcdNode        = KeyEtcdNodeProfile + "%s"

	//key  /pulse/proc/<node_uuid>/<job_id>/<pid>
	KeyEtcdProcProfile     = keyEtcdProfile + "proc/"
	KeyEtcdNodeProcProfile = KeyEtcdProcProfile + "%s/"
	KeyEtcdJobProcProfile  = KeyEtcdNodeProcProfile + "%d/"
	KeyEtcdProc            = KeyEtcdJobProcProfile + "%d"

	//key /pulse/job/<node_uuid>/<job_id>
	KeyEtcdJobProfile = keyEtcdProfile + "job/%s/"
	KeyEtcdJob        = KeyEtcdJobProfile + "%d"

	// key /pulse/once/<jobID>
	KeyEtcdOnceProfile = keyEtcdProfile + "once/"
	KeyEtcdOnce        = KeyEtcdOnceProfile + "%d"

	KeyEtcdLockProfile = keyEtcdProfile + "lock/"
	KeyEtcdLock        = KeyEtcdLockProfile + "%s"

	// key /pulse/system/<node_uuid>
	KeyEtcdSystemProfile = keyEtcdProfile + "system/"
	KeyEtcdSystemSwitch  = KeyEtcdSystemProfile + "switch/" + "%s"
	KeyEtcdSystemGet     = KeyEtcdSystemProfile + "get/" + "%s"
)
