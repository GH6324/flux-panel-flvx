package com.admin.common.dto;

import lombok.Data;

import java.util.ArrayList;
import java.util.List;

@Data
public class TunnelGroupDetailDto {
    private Long id;
    private String name;
    private Integer status;
    private Long createdTime;
    private Long updatedTime;
    private List<Long> tunnelIds = new ArrayList<>();
    private List<String> tunnelNames = new ArrayList<>();
}
