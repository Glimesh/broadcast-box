import React, { useContext, useEffect, useState } from "react";
import TextInputDialog from "../shared/TextInputDialog";
import AdminFrontpage from "./Frontpage";
import { LocaleContext } from "../../providers/LocaleProvider";

const ADMIN_TOKEN = "adminToken";

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
        Authorization: `Bearer ${token}`,
      },
    })
      .then((result) => {
        if (result.status > 400 && result.status < 500) {
          setErrorMessage("Invalid login");
          return;
        }

        return result.json();
      })
      .then((result: LoginResponse) => {
        if (result.isValid) {
          localStorage.setItem(ADMIN_TOKEN, token);
          setIsLoggedIn(() => true);
        } else {
          localStorage.removeItem(ADMIN_TOKEN);
          setErrorMessage(() => result.errorMessage);
        }
      });
  };

  useEffect(() => {
    const token = localStorage.getItem(ADMIN_TOKEN);

    if (!token) {
      return;
    }

    fetch(`/api/admin/login`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
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
        <div className="mt-4 w-full max-w-md mx-auto rounded-xl border border-red-400 bg-gray-800 p-4 shadow-md">
          <div className="flex items-center space-x-3">
            <svg
              className="h-6 w-6 text-red-600"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <h2 className="text-lg font-semibold text-red-700">
              {locale.admin_login.error_message_login_failed}
            </h2>
          </div>
          <p className="mt-2 pl-6 text-sm text-red-600">{errorMessage}</p>
        </div>
      )}
    </div>
  );
};

export default Admin;
