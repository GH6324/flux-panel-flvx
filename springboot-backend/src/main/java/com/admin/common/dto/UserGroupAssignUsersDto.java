package com.admin.common.dto;

import lombok.Data;

import javax.validation.constraints.NotNull;
import java.util.List;

@Data
public class UserGroupAssignUsersDto {

    @NotNull(message = "用户分组ID不能为空")
    private Long groupId;

    @NotNull(message = "用户列表不能为空")
    private List<Long> userIds;
}
