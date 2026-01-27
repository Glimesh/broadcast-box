import React, { useMemo, useState } from "react";
import { localeInterface } from "../locale/localeInterface";
import locale_en from "../locale/en";
import locale_da from "../locale/da";

interface LocaleProviderProps {
	children: React.ReactNode;
}

interface LocaleProviderContextProps {
	locale: localeInterface

	// eslint-disable-next-line no-unused-vars
	setLocale: (locale: LocaleType) => void
}

export const LocaleContext = React.createContext<LocaleProviderContextProps>({
	locale: locale_en,
	// eslint-disable-next-line no-unused-vars
	setLocale: (_: LocaleType) => { }
});

export function LocaleProvider(props: LocaleProviderProps) {
	const initialLocale = localStorage.getItem("locale")
	const [currentLocale, setCurrentLocale] = useState<localeInterface>(getLocaleInterfaceByLocale(initialLocale ?? "en"))

	const state = useMemo<LocaleProviderContextProps>(() => ({
		locale: currentLocale,

		setLocale: (locale: LocaleType) => {
			localStorage.setItem("locale", locale.locale)
			setCurrentLocale(getLocaleInterfaceByLocale(locale.locale))
		}
	}), [currentLocale]);


	return (
		<LocaleContext.Provider value={state}>
			{props.children}
		</LocaleContext.Provider>
	);
}

export type LocaleType = {
	locale: string,
	name: string
}

export const LocaleTypes: LocaleType[] = [
	{
		locale: "en",
		name: "English"
	},
	{
		locale: "dk",
		name: "Dansk"
	},
]

const getLocaleInterfaceByLocale = (locale: string) => {
	switch (locale) {
		case "en":
			return locale_en

		case "dk":
			return locale_da

		default:
			return locale_en
	}
}
