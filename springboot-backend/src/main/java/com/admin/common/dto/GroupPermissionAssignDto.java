package com.admin.common.dto;

import lombok.Data;

import javax.validation.constraints.NotNull;

@Data
public class GroupPermissionAssignDto {

    @NotNull(message = "用户分组ID不能为空")
    private Long userGroupId;

    @NotNull(message = "隧道分组ID不能为空")
    private Long tunnelGroupId;
}
