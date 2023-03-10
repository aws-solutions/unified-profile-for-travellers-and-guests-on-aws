package permissions

import "testing"

/*****************
* PERMISSIONS
*****************/

func TestPermissions(t *testing.T) {
	groups := []string{"g1", "g2"}
	groupWithSuperAdmin := []string{"cloudrack_test_chain", "CLOUDRACK_SUPER_ADMIN"}
	groupsWithPrefix := []string{"cloudrack_group_g1", "cloudrack_g2"}
	groupsWithPrefixForOrgDetection := []string{"cloudrack_1726381926", "cloudrack_group_123456789", "cloudrack_999999999"}

	if !IsSuperAdmin(groupWithSuperAdmin) {
		t.Errorf("[core][permissions] user in groups %+v should be super admin", groupWithSuperAdmin)
	}
	if IsSuperAdmin(groupsWithPrefix) {
		t.Errorf("[core][permissions] user in groups %+v should not be super admin", groupsWithPrefix)
	}
	if !HasGroup(groups, "g1") {
		t.Errorf("[core][permissions] user in groups %+v has group g1", groups)
	}
	if !HasGroup(groups, "g2") {
		t.Errorf("[core][permissions] user in groups %+v has group g2", groups)
	}
	if HasGroup(groups, "g3") {
		t.Errorf("[core][permissions] user in groups %+v does not have has group g3", groups)
	}
	if GroupName("123456789") != "cloudrack_group_"+"123456789" {
		t.Errorf("[core][permissions] Group name should be %s and not %s", "cloudrack_group_"+"123456789", GroupName("123456789"))
	}
	if !HasGroupWithID(groupsWithPrefix, "g1") {
		t.Errorf("[core][permissions] user in groups %+v has group with ID g1", groups)
	}
	if HasGroupWithID(groupsWithPrefix, "g2") {
		t.Errorf("[core][permissions] user in groups %+v does noot have group with ID g2", groupsWithPrefix)
	}
	if GetOrgID(groupsWithPrefixForOrgDetection) != "123456789" {
		t.Errorf("[core][permissions] user in groups %+v shoudl have org ID '123456789'", groupsWithPrefixForOrgDetection)
	}
}
