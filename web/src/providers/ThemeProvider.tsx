import React, { createContext, useCallback, useEffect, useMemo, useState } from "react";

export type ThemePreference = "light" | "dark" | "system";
export type ResolvedTheme = "light" | "dark";

const STORAGE_KEY = "bb_theme";

const THEME_COLORS: Record<ResolvedTheme, string> = {
	light: "#dff8ff",
	dark: "#0b1428",
};

interface ThemeContextProps {
	theme: ThemePreference;
	resolvedTheme: ResolvedTheme;
	setTheme: (theme: ThemePreference) => void;
}

export const ThemeContext = createContext<ThemeContextProps>({
	theme: "light",
	resolvedTheme: "light",
	setTheme: () => {},
});

const getSystemTheme = (): ResolvedTheme => {
	if (typeof window === "undefined") return "light";
	return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
};

const getStoredTheme = (): ThemePreference => {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored === "light" || stored === "dark" || stored === "system") {
			return stored;
		}
	} catch {
		// localStorage may be unavailable in private browsing modes.
	}

	return "light";
};

const applyTheme = (resolvedTheme: ResolvedTheme) => {
	const root = document.documentElement;
	root.dataset.theme = resolvedTheme;
	root.style.colorScheme = resolvedTheme;

	const themeColor = document.querySelector('meta[name="theme-color"]');
	if (themeColor) {
		themeColor.setAttribute("content", THEME_COLORS[resolvedTheme]);
	}
};

export function ThemeProvider({ children }: { children: React.ReactNode }) {
	const [theme, setThemeState] = useState<ThemePreference>(getStoredTheme);
	const [systemTheme, setSystemTheme] = useState<ResolvedTheme>(getSystemTheme);
	const resolvedTheme = theme === "system" ? systemTheme : theme;

	useEffect(() => {
		applyTheme(resolvedTheme);
	}, [resolvedTheme]);

	useEffect(() => {
		const media = window.matchMedia("(prefers-color-scheme: dark)");
		const handleChange = (event: MediaQueryListEvent) => {
			setSystemTheme(event.matches ? "dark" : "light");
		};

		media.addEventListener("change", handleChange);
		return () => media.removeEventListener("change", handleChange);
	}, []);

	const setTheme = useCallback((nextTheme: ThemePreference) => {
		setThemeState(nextTheme);
		try {
			localStorage.setItem(STORAGE_KEY, nextTheme);
		} catch {
			// The preference still applies for this session.
		}
	}, []);

	const state = useMemo<ThemeContextProps>(() => ({
		theme,
		resolvedTheme,
		setTheme,
	}), [theme, resolvedTheme, setTheme]);

	return (
		<ThemeContext.Provider value={state}>
			{children}
		</ThemeContext.Provider>
	);
}
