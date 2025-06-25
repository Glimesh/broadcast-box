import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'
import {BrowserRouter} from "react-router-dom";
import { CinemaModeProvider } from './providers/CinemaModeProvider';
import {StatusProvider} from "./providers/StatusProvider";

// @ts-ignore
const root = ReactDOM.createRoot(document.getElementById('root'))
const path = import.meta.env.PUBLIC_URL;

root.render(
	<React.StrictMode>
		<BrowserRouter basename={path}>
			<StatusProvider>
				<CinemaModeProvider>
					<App/>
				</CinemaModeProvider>
			</StatusProvider>
		</BrowserRouter>
	</React.StrictMode>
)
