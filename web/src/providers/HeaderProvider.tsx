import React, { useMemo, useState } from "react";

interface HeaderProviderContextProps {
	title: string;
	setTitle: (title: string) => void
}

export const HeaderContext = React.createContext<HeaderProviderContextProps>({
	title: "BroadcastBox",
	setTitle: () => { },
});

interface HeaderProviderProps {
	children: React.ReactNode;
}

export function HeaderProvider(props: HeaderProviderProps) {
	const [title, setTitle] = useState("BroadcastBox")

	const state = useMemo<HeaderProviderContextProps>(() => ({
		title: title,
		setTitle: (value) => setTitle(() => value !== "" ? `BroadcastBox - ${value}` : "BroadcastBox")
	}), [title]);

	return (
		<HeaderContext.Provider value={state}>
			{props.children}
		</HeaderContext.Provider>
	);
}
