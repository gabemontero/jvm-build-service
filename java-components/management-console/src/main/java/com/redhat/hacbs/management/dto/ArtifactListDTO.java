package com.redhat.hacbs.management.dto;

import org.eclipse.microprofile.openapi.annotations.media.Schema;

public record ArtifactListDTO(
        @Schema(required = true) String gav,

        @Schema(required = true) String name,

        boolean succeeded, boolean missing, String message) {

}
