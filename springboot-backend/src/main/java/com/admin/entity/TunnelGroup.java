package com.admin.entity;

import lombok.Data;
import lombok.EqualsAndHashCode;

@Data
@EqualsAndHashCode(callSuper = true)
public class TunnelGroup extends BaseEntity {

    private static final long serialVersionUID = 1L;

    private String name;
}
