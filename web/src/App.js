import React from 'react';
import { Routes, Route } from 'react-router-dom';
import RootWrapper from './components/rootWrapper/rootWrapper';
import Frontpage from "./components/selection/frontpage";
import BrowserBroadcaster from "./components/broadcast/Broadcast";
import PlayerPage from "./components/player/PlayerPage";
function App() {
    return (React.createElement(Routes, null,
        React.createElement(Route, { path: '/', element: React.createElement(RootWrapper, null) },
            React.createElement(Route, { index: true, element: React.createElement(Frontpage, null) }),
            React.createElement(Route, { path: '/publish/*', element: React.createElement(BrowserBroadcaster, null) }),
            React.createElement(Route, { path: '/*', element: React.createElement(PlayerPage, null) }))));
}
export default App;
