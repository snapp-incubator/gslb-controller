# Test


1. Create glsb with multiple backends, it must create all gslbcontents
2. Update: Add a new entry in backend, it must add the corresponding gslbcontents
3. Update: Delete an entry in backend, it must delete the corresponding gslbcontents
4. Update: edit an entry in backend, must edit gslbcontent
5. Delete: Delete gslb, it must delete all gslbcontents, and finalize
6. Reconcile: if a gslbcontent is deleted manually it must be recreated again automatically
7. Reconcile: edit a gslbcontent, it should revert manual changes again.



Bugs:

* why required on edit does not work? can delete required fields in CR, cause panic
* duplicate names in backends, cause loop
