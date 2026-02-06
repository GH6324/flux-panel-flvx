package com.admin.controller;

import com.admin.common.aop.LogAnnotation;
import com.admin.common.annotation.RequireRole;
import com.admin.common.dto.GroupCreateDto;
import com.admin.common.dto.GroupPermissionAssignDto;
import com.admin.common.dto.GroupUpdateDto;
import com.admin.common.dto.TunnelGroupAssignTunnelsDto;
import com.admin.common.dto.UserGroupAssignUsersDto;
import com.admin.common.lang.R;
import com.admin.service.GroupService;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import javax.annotation.Resource;
import java.util.Map;

@RestController
@CrossOrigin
@RequestMapping("/api/v1/group")
public class GroupController {

    @Resource
    private GroupService groupService;

    @LogAnnotation
    @RequireRole
    @PostMapping("/tunnel/list")
    public R tunnelGroupList() {
        return groupService.getTunnelGroups();
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/tunnel/create")
    public R createTunnelGroup(@Validated @RequestBody GroupCreateDto dto) {
        return groupService.createTunnelGroup(dto);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/tunnel/update")
    public R updateTunnelGroup(@Validated @RequestBody GroupUpdateDto dto) {
        return groupService.updateTunnelGroup(dto);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/tunnel/delete")
    public R deleteTunnelGroup(@RequestBody Map<String, Object> params) {
        Long id = Long.valueOf(params.get("id").toString());
        return groupService.deleteTunnelGroup(id);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/tunnel/assign")
    public R assignTunnels(@Validated @RequestBody TunnelGroupAssignTunnelsDto dto) {
        return groupService.assignTunnelsToGroup(dto);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/user/list")
    public R userGroupList() {
        return groupService.getUserGroups();
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/user/create")
    public R createUserGroup(@Validated @RequestBody GroupCreateDto dto) {
        return groupService.createUserGroup(dto);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/user/update")
    public R updateUserGroup(@Validated @RequestBody GroupUpdateDto dto) {
        return groupService.updateUserGroup(dto);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/user/delete")
    public R deleteUserGroup(@RequestBody Map<String, Object> params) {
        Long id = Long.valueOf(params.get("id").toString());
        return groupService.deleteUserGroup(id);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/user/assign")
    public R assignUsers(@Validated @RequestBody UserGroupAssignUsersDto dto) {
        return groupService.assignUsersToGroup(dto);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/permission/list")
    public R listPermissions() {
        return groupService.getGroupPermissions();
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/permission/assign")
    public R assignPermission(@Validated @RequestBody GroupPermissionAssignDto dto) {
        return groupService.assignGroupPermission(dto);
    }

    @LogAnnotation
    @RequireRole
    @PostMapping("/permission/remove")
    public R removePermission(@RequestBody Map<String, Object> params) {
        Long id = Long.valueOf(params.get("id").toString());
        return groupService.removeGroupPermission(id);
    }
}
