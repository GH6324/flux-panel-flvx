package com.admin.common.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;

@Data
public class GroupCreateDto {

    @NotBlank(message = "分组名称不能为空")
    private String name;

    private Integer status;
}
