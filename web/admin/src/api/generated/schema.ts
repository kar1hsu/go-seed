// Code generated from the admin .api contract. DO NOT EDIT.
export interface paths {
    "/admin-api/v1/auth/login": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Sign in; returns access token and sets HttpOnly refresh cookie */
        post: operations["authLogin"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/auth/logout": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Revoke current session and clear refresh cookie */
        post: operations["authLogout"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/auth/me": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Read the current administrator; checks live session state */
        get: operations["authCurrent"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/auth/password": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Change password and revoke every administrator session */
        post: operations["authChangePassword"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/auth/refresh": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Rotate HttpOnly refresh cookie and access token */
        post: operations["authRefresh"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/administrators": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** ListAdministrators */
        get: operations["managementListAdministrators"];
        put?: never;
        /** CreateAdministrator */
        post: operations["managementCreateAdministrator"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/administrators/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** UpdateAdministrator */
        patch: operations["managementUpdateAdministrator"];
        trace?: never;
    };
    "/admin-api/v1/system/administrators/{id}/roles": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** AssignAdministratorRoles */
        put: operations["managementAssignAdministratorRoles"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/audit-logs": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** ListAuditLogs */
        get: operations["auditListAuditLogs"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/audit-logs/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** GetAuditLog */
        get: operations["auditGetAuditLog"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/dictionaries": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** ListDictionaries */
        get: operations["settingsListDictionaries"];
        put?: never;
        /** CreateDictionary */
        post: operations["settingsCreateDictionary"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/dictionaries/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** GetDictionary */
        get: operations["settingsGetDictionary"];
        /** UpdateDictionary */
        put: operations["settingsUpdateDictionary"];
        post?: never;
        /** DeleteDictionary */
        delete: operations["settingsDeleteDictionary"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/permissions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** ListPermissions */
        get: operations["managementListPermissions"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/ping": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Probe API to system-rpc connectivity */
        get: operations["systemPing"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/roles": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** ListRoles */
        get: operations["managementListRoles"];
        put?: never;
        /** CreateRole */
        post: operations["managementCreateRole"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/roles/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** DeleteRole */
        delete: operations["managementDeleteRole"];
        options?: never;
        head?: never;
        /** UpdateRole */
        patch: operations["managementUpdateRole"];
        trace?: never;
    };
    "/admin-api/v1/system/roles/{id}/permissions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** GrantRole */
        put: operations["managementGrantRole"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/settings": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** ListSettings */
        get: operations["settingsListSettings"];
        put?: never;
        /** CreateSetting */
        post: operations["settingsCreateSetting"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/admin-api/v1/system/settings/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** UpdateSetting */
        put: operations["settingsUpdateSetting"];
        post?: never;
        /** DeleteSetting */
        delete: operations["settingsDeleteSetting"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        ActionData: {
            success: boolean;
        };
        ActionResponse: {
            code: number;
            data: components["schemas"]["ActionData"];
            message: string;
            traceId: string;
        };
        AdminProfile: {
            displayName: string;
            id: string;
            username: string;
        };
        AdministratorListData: {
            items: components["schemas"]["AdministratorRecord"][];
            page: number;
            pageSize: number;
            total: number;
        };
        AdministratorListResponse: {
            code: number;
            data: components["schemas"]["AdministratorListData"];
            message: string;
            traceId: string;
        };
        AdministratorRecord: {
            createdAt: string;
            displayName: string;
            enabled: boolean;
            id: string;
            manageable: boolean;
            revision: string;
            roleIds: string[];
            roleNames: string[];
            username: string;
        };
        AdministratorResponse: {
            code: number;
            data: components["schemas"]["AdministratorRecord"];
            message: string;
            traceId: string;
        };
        AssignAdministratorRolesRequest: {
            revision: string;
            roleIds: string[];
        };
        AuditAction: {
            code: string;
            name: string;
        };
        AuditChange: {
            after: string[];
            before: string[];
            field: string;
            label: string;
        };
        AuditLogDetailData: {
            changes: components["schemas"]["AuditChange"][];
            record: components["schemas"]["AuditLogRecord"];
        };
        AuditLogDetailResponse: {
            code: number;
            data: components["schemas"]["AuditLogDetailData"];
            message: string;
            traceId: string;
        };
        AuditLogListData: {
            actions: components["schemas"]["AuditAction"][];
            items: components["schemas"]["AuditLogRecord"][];
            page: number;
            pageSize: number;
            total: number;
        };
        AuditLogListResponse: {
            code: number;
            data: components["schemas"]["AuditLogListData"];
            message: string;
            traceId: string;
        };
        AuditLogRecord: {
            action: string;
            actionName: string;
            actorId: string;
            actorName: string;
            actorUsername: string;
            createdAt: string;
            id: string;
            source: string;
            targetId: string;
            targetKey: string;
            targetName: string;
            targetType: string;
            traceId: string;
        };
        AuthData: {
            accessToken: string;
            admin: components["schemas"]["AdminProfile"];
            expiresAt: string;
        };
        AuthResponse: {
            code: number;
            data: components["schemas"]["AuthData"];
            message: string;
            traceId: string;
        };
        ChangePasswordRequest: {
            currentPassword: string;
            newPassword: string;
        };
        CreateAdministratorRequest: {
            displayName: string;
            password: string;
            username: string;
        };
        CreateDictionaryRequest: {
            code: string;
            description: string;
            enabled: boolean;
            items: components["schemas"]["DictionaryItem"][];
            name: string;
        };
        CreateRoleRequest: {
            code: string;
            description: string;
            name: string;
        };
        CreateSettingRequest: {
            code: string;
            description: string;
            enabled: boolean;
            name: string;
            type: string;
            value: string;
        };
        CurrentData: {
            admin: components["schemas"]["AdminProfile"];
            permissions: string[];
            policyVersion: string;
            roles: string[];
            superAdmin: boolean;
        };
        CurrentResponse: {
            code: number;
            data: components["schemas"]["CurrentData"];
            message: string;
            traceId: string;
        };
        DeleteRoleRequest: Record<string, never>;
        DeleteSettingResourceRequest: Record<string, never>;
        DictionaryDetailData: {
            items: components["schemas"]["DictionaryItem"][];
            record: components["schemas"]["DictionaryRecord"];
        };
        DictionaryItem: {
            enabled: boolean;
            label: string;
            sort: number;
            value: string;
        };
        DictionaryListData: {
            items: components["schemas"]["DictionaryRecord"][];
            page: number;
            pageSize: number;
            total: number;
        };
        DictionaryListResponse: {
            code: number;
            data: components["schemas"]["DictionaryListData"];
            message: string;
            traceId: string;
        };
        DictionaryRecord: {
            code: string;
            createdAt: string;
            description: string;
            enabled: boolean;
            id: string;
            itemCount: number;
            name: string;
            revision: string;
        };
        DictionaryResponse: {
            code: number;
            data: components["schemas"]["DictionaryDetailData"];
            message: string;
            traceId: string;
        };
        GetAuditLogRequest: Record<string, never>;
        GetDictionaryRequest: Record<string, never>;
        GrantRoleRequest: {
            permissions: string[];
            revision: string;
        };
        ListAuditLogsRequest: Record<string, never>;
        ListManagementRequest: Record<string, never>;
        LoginRequest: {
            password: string;
            username: string;
        };
        PermissionListData: {
            items: components["schemas"]["PermissionRecord"][];
        };
        PermissionListResponse: {
            code: number;
            data: components["schemas"]["PermissionListData"];
            message: string;
            traceId: string;
        };
        PermissionRecord: {
            code: string;
            group: string;
            kind: string;
            name: string;
            requires: string[];
        };
        PingData: {
            service: string;
            time: string;
            version: string;
        };
        PingResponse: {
            code: number;
            data: components["schemas"]["PingData"];
            message: string;
            traceId: string;
        };
        RoleListData: {
            items: components["schemas"]["RoleRecord"][];
            page: number;
            pageSize: number;
            total: number;
        };
        RoleListResponse: {
            code: number;
            data: components["schemas"]["RoleListData"];
            message: string;
            traceId: string;
        };
        RoleRecord: {
            assignable: boolean;
            builtIn: boolean;
            code: string;
            description: string;
            enabled: boolean;
            id: string;
            manageable: boolean;
            memberCount: number;
            name: string;
            permissions: string[];
            revision: string;
        };
        RoleResponse: {
            code: number;
            data: components["schemas"]["RoleRecord"];
            message: string;
            traceId: string;
        };
        SettingListData: {
            items: components["schemas"]["SettingRecord"][];
            page: number;
            pageSize: number;
            total: number;
        };
        SettingListResponse: {
            code: number;
            data: components["schemas"]["SettingListData"];
            message: string;
            traceId: string;
        };
        SettingRecord: {
            code: string;
            createdAt: string;
            description: string;
            enabled: boolean;
            id: string;
            name: string;
            revision: string;
            type: string;
            value: string;
        };
        SettingResponse: {
            code: number;
            data: components["schemas"]["SettingRecord"];
            message: string;
            traceId: string;
        };
        UpdateAdministratorRequest: {
            displayName: string;
            enabled: boolean;
            revision: string;
        };
        UpdateDictionaryRequest: {
            description: string;
            enabled: boolean;
            items: components["schemas"]["DictionaryItem"][];
            name: string;
            revision: string;
        };
        UpdateRoleRequest: {
            description: string;
            enabled: boolean;
            name: string;
            revision: string;
        };
        UpdateSettingRequest: {
            description: string;
            enabled: boolean;
            name: string;
            revision: string;
            value: string;
        };
        ErrorResponse: {
            code: number;
            message: string;
            /** @enum {unknown|null} */
            data: null;
            traceId: string;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    authLogin: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["LoginRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    /** @description Host-only goseed_refresh cookie; HttpOnly; SameSite=Strict; Path=/admin-api/v1/auth; Secure over HTTPS. */
                    "Set-Cookie"?: string;
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AuthResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    authLogout: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    authCurrent: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["CurrentResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    authChangePassword: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ChangePasswordRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    authRefresh: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    /** @description Host-only goseed_refresh cookie; HttpOnly; SameSite=Strict; Path=/admin-api/v1/auth; Secure over HTTPS. */
                    "Set-Cookie"?: string;
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AuthResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementListAdministrators: {
        parameters: {
            query: {
                /** @example 1 */
                page: number;
                /** @example 20 */
                pageSize: number;
                keyword?: string;
                /** @example all */
                enabled: "all" | "enabled" | "disabled";
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AdministratorListResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementCreateAdministrator: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CreateAdministratorRequest"];
            };
        };
        responses: {
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AdministratorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementUpdateAdministrator: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateAdministratorRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementAssignAdministratorRoles: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["AssignAdministratorRolesRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    auditListAuditLogs: {
        parameters: {
            query: {
                /** @example 1 */
                page: number;
                /** @example 20 */
                pageSize: number;
                keyword?: string;
                action?: string;
                actorId?: string;
                traceId?: string;
                /** @description Inclusive RFC3339 timestamp. */
                from?: string;
                /** @description Exclusive RFC3339 timestamp. */
                to?: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AuditLogListResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    auditGetAuditLog: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AuditLogDetailResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsListDictionaries: {
        parameters: {
            query: {
                /** @example 1 */
                page: number;
                /** @example 20 */
                pageSize: number;
                keyword?: string;
                /** @example all */
                enabled: "all" | "enabled" | "disabled";
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DictionaryListResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsCreateDictionary: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CreateDictionaryRequest"];
            };
        };
        responses: {
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DictionaryResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsGetDictionary: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DictionaryResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsUpdateDictionary: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateDictionaryRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsDeleteDictionary: {
        parameters: {
            query: {
                revision: string;
            };
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementListPermissions: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PermissionListResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    systemPing: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PingResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementListRoles: {
        parameters: {
            query: {
                /** @example 1 */
                page: number;
                /** @example 20 */
                pageSize: number;
                keyword?: string;
                /** @example all */
                enabled: "all" | "enabled" | "disabled";
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RoleListResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementCreateRole: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CreateRoleRequest"];
            };
        };
        responses: {
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RoleResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementDeleteRole: {
        parameters: {
            query: {
                revision: string;
            };
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementUpdateRole: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateRoleRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    managementGrantRole: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["GrantRoleRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsListSettings: {
        parameters: {
            query: {
                /** @example 1 */
                page: number;
                /** @example 20 */
                pageSize: number;
                keyword?: string;
                /** @example all */
                enabled: "all" | "enabled" | "disabled";
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SettingListResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsCreateSetting: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CreateSettingRequest"];
            };
        };
        responses: {
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SettingResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsUpdateSetting: {
        parameters: {
            query?: never;
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateSettingRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
    settingsDeleteSetting: {
        parameters: {
            query: {
                revision: string;
            };
            header: {
                /** @description Required for same-origin write requests, including CLI clients. */
                "X-GoSeed-CSRF": "1";
            };
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ActionResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            404: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            408: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            429: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            503: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
            /** @description GoSeed error envelope; code is stable, message is safe for clients. */
            504: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ErrorResponse"];
                };
            };
        };
    };
}
