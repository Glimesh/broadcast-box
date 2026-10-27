import { useContext, useState } from "react";
import { ThemeContext, ThemePreference } from "../../providers/ThemeProvider";
import Button from "./Button";
import { getIcon, IconType } from "./Icons";

interface ThemeOption {
	value: ThemePreference;
	title: string;
	icon?: IconType;
}

const themeOptions: ThemeOption[] = [
	{ value: "light", title: "light ♡", icon: "Sun" },
	{ value: "dark", title: "dark ☾", icon: "Moon" },
	{ value: "system", title: "system", icon: "Auto" },
];

export default function ThemeSelector() {
	const { theme, resolvedTheme, setTheme } = useContext(ThemeContext);
	const [isOpen, setIsOpen] = useState(false);

	return (
		<div className="relative">
			<button
				type="button"
				className="nav-link theme-toggle"
				onClick={() => setIsOpen((prev) => !prev)}
				aria-label="change theme"
				aria-haspopup="true"
				aria-expanded={isOpen}
			>
				<span className="theme-toggle__label">theme</span>
				{getIcon(resolvedTheme === "dark" ? "Moon" : "Sun")}
			</button>

			{isOpen && (
				<>
					<div
						className="fixed inset-0 z-40"
						onClick={() => setIsOpen(false)}
					/>
					<div className="window theme-menu absolute right-0 top-[calc(100%+8px)] z-50 flex w-48 flex-col gap-1 p-2 animate-scale-in">
						{themeOptions.map((option) => (
							<Button
								key={option.value}
								title={`${option.title}${theme === option.value ? " ✓" : ""}`}
								iconLeft={option.icon}
								stretch
								classNames={`justify-start theme-option ${theme === option.value ? "is-selected" : ""}`}
								onClick={() => {
									setTheme(option.value);
									setIsOpen(false);
								}}
							/>
						))}
						<p className="theme-menu__note">system follows your device</p>
					</div>
				</>
			)}
		</div>
	);
}
