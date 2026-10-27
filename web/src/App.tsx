import { Routes, Route } from "react-router-dom";

import BrowserBroadcaster from "./components/broadcast/Broadcast";
import PlayerPage from "./components/player/PlayerPage";
import RootWrapper from "./components/rootWrapper/RootWrapper";
import Frontpage from "./components/selection/Frontpage";
import Statistics from "./components/statistics/Statistics";
import Admin from "./components/admin/Login";
import AuthPage from "./components/account/AuthPage";
import AccountPage from "./components/account/AccountPage";
import ProfilePage from "./components/account/ProfilePage";

function App() {
  return (
    <Routes>
      <Route path="/" element={<RootWrapper />}>
        <Route index element={<Frontpage />} />
        <Route path="/statistics" element={<Statistics />} />
        <Route path="/login" element={<AuthPage />} />
        <Route path="/account" element={<AccountPage />} />
        <Route path="/u/:username" element={<ProfilePage />} />
        <Route path="/publish/*" element={<BrowserBroadcaster />} />
        <Route path="/admin/*" element={<Admin />} />
        <Route path="/*" element={<PlayerPage />} />
      </Route>
    </Routes>
  );
}

export default App;
