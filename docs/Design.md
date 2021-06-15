# Design

- what if duplicate name in bakends? check for it in creation? yes
- service name MUST be unique in all namesapces (ingress URLs)
- service names can be changed, so cannot be used in ConsulService name, UID must be used
- What if some user deletes finalizer, or foce delete the object? shouldn't we have a cluster-scope object representing the external entity?
- what if automatic provisioning does not happen (migration, etc), and a manually admin-created content needs to be refrenced by the CR, and not provisioning a new one (also note new CR has a new UID)
- add services to: prefixName-uid-backend.name
