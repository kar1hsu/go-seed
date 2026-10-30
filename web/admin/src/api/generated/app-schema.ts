// Code generated from the app .api contract. DO NOT EDIT.
export interface paths {
    "/api/v1/auth/login": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Sign in with member credentials; sets a member-only HttpOnly cookie */
        post: operations["memberMemberLogin"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/logout": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** MemberLogout */
        post: operations["memberMemberLogout"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/refresh": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** MemberRefresh */
        post: operations["memberMemberRefresh"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/auth/register": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Register a member account; sign in separately */
        post: operations["memberRegisterMember"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/me": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Read the current member; no client-supplied user ID */
        get: operations["memberMemberCurrent"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** UpdateMemberProfile */
        patch: operations["memberUpdateMemberProfile"];
        trace?: never;
    };
    "/api/v1/me/password": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Change password and revoke all member sessions */
        post: operations["memberMemberChangePassword"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/v1/system/ping": {
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
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        MemberActionData: {
            success: boolean;
        };
        MemberActionResponse: {
            code: number;
            data: components["schemas"]["MemberActionData"];
            message: string;
            traceId: string;
        };
        MemberAuthData: {
            accessToken: string;
            expiresAt: string;
            user: components["schemas"]["MemberProfile"];
        };
        MemberAuthResponse: {
            code: number;
            data: components["schemas"]["MemberAuthData"];
            message: string;
            traceId: string;
        };
        MemberLoginRequest: {
            password: string;
            username: string;
        };
        MemberPasswordRequest: {
            currentPassword: string;
            newPassword: string;
        };
        MemberProfile: {
            createdAt: string;
            id: string;
            nickname: string;
            revision: string;
            username: string;
        };
        MemberProfileResponse: {
            code: number;
            data: components["schemas"]["MemberProfile"];
            message: string;
            traceId: string;
        };
        MemberRefreshRequest: Record<string, never>;
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
        RegisterMemberRequest: {
            nickname: string;
            password: string;
            username: string;
        };
        UpdateMemberProfileRequest: {
            nickname: string;
            revision: string;
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
    memberMemberLogin: {
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
                "application/json": components["schemas"]["MemberLoginRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    /** @description Host-only goseed_member_refresh cookie; HttpOnly; SameSite=Strict; Path=/api/v1/auth; Secure over HTTPS. */
                    "Set-Cookie"?: string;
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MemberAuthResponse"];
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
    memberMemberLogout: {
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
                    "application/json": components["schemas"]["MemberActionResponse"];
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
    memberMemberRefresh: {
        parameters: {
            query?: never;
            header: {
                "X-GoSeed-Refresh-Key"?: string;
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
                    /** @description Host-only goseed_member_refresh cookie; HttpOnly; SameSite=Strict; Path=/api/v1/auth; Secure over HTTPS. */
                    "Set-Cookie"?: string;
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MemberAuthResponse"];
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
    memberRegisterMember: {
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
                "application/json": components["schemas"]["RegisterMemberRequest"];
            };
        };
        responses: {
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MemberProfileResponse"];
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
    memberMemberCurrent: {
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
                    "application/json": components["schemas"]["MemberProfileResponse"];
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
    memberUpdateMemberProfile: {
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
                "application/json": components["schemas"]["UpdateMemberProfileRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MemberProfileResponse"];
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
    memberMemberChangePassword: {
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
                "application/json": components["schemas"]["MemberPasswordRequest"];
            };
        };
        responses: {
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MemberActionResponse"];
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
}
