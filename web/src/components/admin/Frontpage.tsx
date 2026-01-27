import React, { useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import Card from "../shared/Card";
import Button from "../shared/Button";
import StatusPage from "./menus/StatusPage";
import ProfilesPage from "./menus/ProfilesPage";
import ApiPage from "./menus/ApiPage";
import LoggingPage from "./menus/LoggingPage";
import { LocaleContext } from "../../providers/LocaleProvider";

const ADMIN_TOKEN = "adminToken";
const AdminFrontpage = () => {
  const navigation = useNavigate();
  const { locale } = useContext(LocaleContext)
  const [currentMenu, setCurrentMenu] = useState<string>("Status");

  const onChangeMenu = (name: string) => setCurrentMenu(() => name);
  const logout = () => {
    localStorage.removeItem(ADMIN_TOKEN);
    navigation("/");
  };

  return (
    <div className="flex items-center h-full w-full flex-col justify-between">
      <h2 className="text-4xl mb-2">{locale.admin_page.title}</h2>

      <div className="flex flex-row w-full h-full space-x-2">
        <div className="flex-1 space-x-2">
          <Card title={locale.admin_page.menu_settings}>
            <div className="flex flex-col h-full p-2 space-y-2">
              <Button title={locale.admin_page.menu_status} stretch onClick={() => onChangeMenu("Status")} />
              <Button title={locale.admin_page.menu_profiles} stretch onClick={() => onChangeMenu("Profiles")} />
              <Button title={locale.admin_page.menu_api} stretch onClick={() => onChangeMenu("API")} isDisabled />
              <Button title={locale.admin_page.menu_logging} stretch onClick={() => onChangeMenu("Logging")} />
            </div>

            <div className="flex flex-col p-2 space-y-2">
              <Button title={locale.admin_page.menu_settings} stretch onClick={() => onChangeMenu("Settings")} isDisabled />
              <Button title={locale.admin_page.menu_logout} stretch onClick={logout} />
            </div>
          </Card>
        </div>

        <div className={"flex-3"}>
          <Card>
            {currentMenu === "Status" && <StatusPage />}
            {currentMenu === "Profiles" && <ProfilesPage />}
            {currentMenu === "API" && <ApiPage />}
            {currentMenu === "Logging" && <LoggingPage />}
          </Card>
        </div>
      </div>
    </div>
  );
};

export default AdminFrontpage;
