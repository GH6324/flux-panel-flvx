package com.admin.common.dto;

import lombok.Data;

import javax.validation.constraints.NotNull;
import java.util.List;

@Data
public class TunnelGroupAssignTunnelsDto {

    @NotNull(message = "隧道分组ID不能为空")
    private Long groupId;

    @NotNull(message = "隧道列表不能为空")
    private List<Long> tunnelIds;
}
