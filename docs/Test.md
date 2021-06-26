# Test


1. Create glsb with multiple backends, it must create all gslbcontents
2. Update: Add a new entry in backend, it must add the corresponding gslbcontents
3. Update: Delete an entry in backend, it must delete the corresponding gslbcontents
4. Update: edit an entry in backend, must edit gslbcontent
5. Delete: Delete gslb, it must delete all gslbcontents, and finalize
6. Reconcile: if a gslbcontent is deleted manually it must be recreated again automatically
7. Reconcile: edit a gslbcontent, it should revert manual changes again.


## Validation

1. duplicate backend name in one Gslb (on update, create)
2. duplicate serviceName among two Gslb (on update, create)
3. delete a serviceName should release it.
4. rename a serviceName should add new, and release old.


## Reliability

1. take the controller down, when resources are already up.. nothing should break
2. take the controller down, add/remove/edit resources: should not be possible
3. take the controller down, even if manually edited (e.g. etcd, make validation: Ignore for some minutes): should not affect crazy and only services with issues (which e.g. uniqueness has been violated, etc) should be affected, not others.
