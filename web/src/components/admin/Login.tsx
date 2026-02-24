import { useContext, useEffect, useState } from "react";
import TextInputDialog from "../shared/TextInputDialog";
import ErrorMessagePanel from "../shared/ErrorMessagePanel";
import AdminFrontpage from "./Frontpage";
import { LocaleContext } from "../../providers/LocaleProvider";
import {
  ADMIN_TOKEN_STORAGE_KEY,
  clearAdminToken,
  getAdminAuthorizationHeader,
  isInvalidAdminSessionResponse,
} from "./adminAuth";

interface LoginResponse {
  isValid: boolean;
  errorMessage: string;
}

const Admin = () => {
  const { locale } = useContext(LocaleContext)
  const [errorMessage, setErrorMessage] = useState<string>("");
  const [isLoggedIn, setIsLoggedIn] = useState<boolean>(false);

  const login = (token: string) => {
    fetch(`/api/admin/login`, {
      method: "POST",
      headers: {
        Authorization: getAdminAuthorizationHeader(token),
      },
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          setErrorMessage("Invalid login");
          return;
        }

        return result.json();
      })
      .then((result: LoginResponse) => {
        if (result.isValid) {
          localStorage.setItem(ADMIN_TOKEN_STORAGE_KEY, token);
          setIsLoggedIn(() => true);
        } else {
          clearAdminToken();
          setErrorMessage(() => result.errorMessage);
        }
      });
  };

  useEffect(() => {
    const token = localStorage.getItem(ADMIN_TOKEN_STORAGE_KEY);

    if (!token) {
      return;
    }

    fetch(`/api/admin/login`, {
      method: "POST",
      headers: {
        Authorization: getAdminAuthorizationHeader(token),
      },
    })
      .then((result) => result.json())
      .then((result: LoginResponse) => {
        if (result.isValid) {
          setIsLoggedIn(true);
        }
      });
  }, []);

  if (isLoggedIn) {
    return <AdminFrontpage />;
  }

  return (
    <div className="flex items-center justify-center h-full w-full flex-col">
      <TextInputDialog<string>
        isSecret={true}
        title={locale.admin_login.login_input_dialog_title}
        message={locale.admin_login.login_input_dialog_message}
        placeholder={locale.admin_login.login_input_dialog_placeholder}
        canCloseOnBackgroundClick={false}
        onAccept={(result: string) => login(result)}
        buttonAcceptText={locale.admin_login.button_login_text}
      />

      {errorMessage !== "" && (
        <ErrorMessagePanel
          title={locale.admin_login.error_message_login_failed}
          message={errorMessage}
        />
      )}
    </div>
  );
};

export default Admin;
