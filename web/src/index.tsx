import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'
import {CinemaModeProvider} from "./components/player/CinemaModeProvider";
import {BrowserRouter} from "react-router-dom";

// @ts-ignore
const root = ReactDOM.createRoot(document.getElementById('root'))
const path = import.meta.env.PUBLIC_URL;

root.render(
  <React.StrictMode>
    <BrowserRouter basename={path}>
      <CinemaModeProvider>
        <App />
      </CinemaModeProvider>
    </BrowserRouter>
  </React.StrictMode>
)
