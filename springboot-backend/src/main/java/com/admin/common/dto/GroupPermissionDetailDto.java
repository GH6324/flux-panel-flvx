package com.admin.common.dto;

import lombok.Data;

@Data
public class GroupPermissionDetailDto {
    private Long id;
    private Long userGroupId;
    private String userGroupName;
    private Long tunnelGroupId;
    private String tunnelGroupName;
    private Long createdTime;
}
