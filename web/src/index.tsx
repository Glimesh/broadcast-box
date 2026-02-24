import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'
import { BrowserRouter } from "react-router-dom";
import { CinemaModeProvider } from './providers/CinemaModeProvider';
import { StatusProvider } from "./providers/StatusProvider";
import { HeaderProvider } from './providers/HeaderProvider';
import { LocaleProvider } from './providers/LocaleProvider';

const rootElement = document.getElementById('root')
if (rootElement === null) {
	throw new Error("Missing root element (#root)")
}

const root = ReactDOM.createRoot(rootElement)
const path = import.meta.env.PUBLIC_URL;

root.render(
	<React.StrictMode>
		<BrowserRouter basename={path}>
			<LocaleProvider>
				<StatusProvider>
					<CinemaModeProvider>
						<HeaderProvider>
							<App />
						</HeaderProvider>
					</CinemaModeProvider>
				</StatusProvider>
			</LocaleProvider>
		</BrowserRouter>
	</React.StrictMode>
)
