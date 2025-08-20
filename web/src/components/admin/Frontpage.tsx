import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Card from "../shared/Card";
import Button from "../shared/Button";
import StatusPage from "./menus/StatusPage";

const ADMIN_TOKEN = "adminToken";
const AdminFrontpage = () => {
  const navigation = useNavigate();
  const [currentMenu, setCurrentMenu] = useState<string>("Status");

  const onChangeMenu = (name: string) => setCurrentMenu(() => name);
  const logout = () => {
    localStorage.removeItem(ADMIN_TOKEN);
    navigation("/");
  };

  return (
    <div className="flex items-center h-full w-full flex-col justify-between">
      <h2 className="text-4xl mb-2">ADMIN PORTAL</h2>

      <div className="flex flex-row w-full h-full space-x-2">
        <div className="flex-1 space-x-2">
          <Card title={"Settings"}>
            <div className="flex flex-col h-full p-2 space-y-2">
              <Button title="Status" onClick={() => onChangeMenu("Status")} />
              <Button title="Profiles" onClick={() => onChangeMenu("Profiles")} />
              <Button title="API" onClick={() => onChangeMenu("API")} />
              <Button title="Logging" onClick={() => onChangeMenu("Logging")} />
            </div>

            <div className="flex flex-col p-2 space-y-2">
              <Button title="Settings" onClick={() => onChangeMenu("Settings")} />
              <Button title="Log out" onClick={logout} />
            </div>
          </Card>
        </div>

        <div className={"flex-3 h-full"}>
          <Card title={currentMenu}>
            {currentMenu === "Status" && <StatusPage />}
            {currentMenu !== "Status" && <div>Not implemented</div>}
          </Card>
        </div>
      </div>
    </div>
  );
};

export default AdminFrontpage;
