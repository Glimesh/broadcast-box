import { useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import Button from "../shared/Button";
import StatusPage from "./menus/StatusPage";
import ProfilesPage from "./menus/ProfilesPage";
import UsersPage from "./menus/UsersPage";
import MusicPage from "./menus/MusicPage";
import ApiPage from "./menus/ApiPage";
import LoggingPage from "./menus/LoggingPage";
import EmotesPage from "./menus/EmotesPage";
import { LocaleContext } from "../../providers/LocaleProvider";
import { clearAdminToken } from "./adminAuth";

const AdminFrontpage = () => {
  const navigation = useNavigate();
  const { locale } = useContext(LocaleContext);
  const [currentMenu, setCurrentMenu] = useState<string>("Status");

  const onChangeMenu = (name: string) => setCurrentMenu(() => name);
  const logout = () => {
    clearAdminToken();
    navigation("/");
  };

  return (
    <div className="page-container animate-slide-up">
      <div className="section-head mt-0">
        <div className="section-head__left">
          <span className="section-kicker">admin dashboard</span>
          <h2>{locale.admin_page.title}</h2>
        </div>
        <div className="section-head__right">be nice ♡</div>
      </div>

      <div className="admin-layout">
        <div className="admin-sidebar">
          <div className="window">
            <div className="window-bar window-bar--hot">
              <span className="window-dots" aria-hidden="true">
                <span className="window-dot" />
                <span className="window-dot" />
                <span className="window-dot" />
              </span>
              <span>menu.cfg</span>
            </div>
            <div className="window-body">
              <div className="admin-menu">
                <Button title={locale.admin_page.menu_status} stretch onClick={() => onChangeMenu("Status")} />
                <Button title={locale.admin_page.menu_profiles} stretch onClick={() => onChangeMenu("Profiles")} />
                <Button title="Users" stretch onClick={() => onChangeMenu("Users")} />
                <Button title="Music" stretch onClick={() => onChangeMenu("Music")} />
                <Button title="Emotes" stretch onClick={() => onChangeMenu("Emotes")} />
                <Button title={locale.admin_page.menu_api} stretch onClick={() => onChangeMenu("API")} isDisabled />
                <Button title={locale.admin_page.menu_logging} stretch onClick={() => onChangeMenu("Logging")} />
                <Button title={locale.admin_page.menu_settings} stretch onClick={() => onChangeMenu("Settings")} isDisabled />
                <Button title={locale.admin_page.menu_logout} color="Deny" stretch onClick={logout} />
              </div>
            </div>
          </div>
        </div>

        <div className="admin-content">
          <div className="window">
            <div className="window-bar">
              <span className="window-dots" aria-hidden="true">
                <span className="window-dot" />
                <span className="window-dot" />
                <span className="window-dot" />
              </span>
              <span>admin ♡ {currentMenu}</span>
              <span className="window-bar__spacer" />
              <span className="window-bar__code">welcome back</span>
            </div>
            <div className="window-body">
              {currentMenu === "Status" && <StatusPage />}
              {currentMenu === "Profiles" && <ProfilesPage />}
              {currentMenu === "Users" && <UsersPage />}
              {currentMenu === "Music" && <MusicPage />}
              {currentMenu === "Emotes" && <EmotesPage />}
              {currentMenu === "API" && <ApiPage />}
              {currentMenu === "Logging" && <LoggingPage />}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminFrontpage;
