import toBase64Utf8 from "../../utilities/base64";

export const ADMIN_TOKEN_STORAGE_KEY = "adminToken";

export const clearAdminToken = () => {
	localStorage.removeItem(ADMIN_TOKEN_STORAGE_KEY);
};

export const getAdminAuthorizationHeader = (token?: string | null) => {
	return `Bearer ${toBase64Utf8(token ?? localStorage.getItem(ADMIN_TOKEN_STORAGE_KEY))}`;
};

export const isInvalidAdminSessionResponse = (status: number) => {
	return status > 400 && status < 500;
};
