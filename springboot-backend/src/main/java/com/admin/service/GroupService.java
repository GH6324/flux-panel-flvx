package com.admin.service;

import com.admin.common.dto.GroupCreateDto;
import com.admin.common.dto.GroupPermissionAssignDto;
import com.admin.common.dto.GroupUpdateDto;
import com.admin.common.dto.TunnelGroupAssignTunnelsDto;
import com.admin.common.dto.UserGroupAssignUsersDto;
import com.admin.common.lang.R;

public interface GroupService {

    R getTunnelGroups();

    R createTunnelGroup(GroupCreateDto dto);

    R updateTunnelGroup(GroupUpdateDto dto);

    R deleteTunnelGroup(Long id);

    R assignTunnelsToGroup(TunnelGroupAssignTunnelsDto dto);

    R getUserGroups();

    R createUserGroup(GroupCreateDto dto);

    R updateUserGroup(GroupUpdateDto dto);

    R deleteUserGroup(Long id);

    R assignUsersToGroup(UserGroupAssignUsersDto dto);

    R getGroupPermissions();

    R assignGroupPermission(GroupPermissionAssignDto dto);

    R removeGroupPermission(Long id);
}
