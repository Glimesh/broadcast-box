import { useContext } from 'react';
import { Link, Outlet } from 'react-router-dom';
import { CinemaModeContext } from "../player/CinemaModeProvider";
import React from 'react';
const RootWrapper = () => {
    const { cinemaMode } = useContext(CinemaModeContext);
    const navbarEnabled = !cinemaMode;
    return (React.createElement("div", null,
        navbarEnabled && (React.createElement("nav", { className: 'bg-gray-800 p-2 mt-0 fixed w-full z-10 top-0' },
            React.createElement("div", { className: 'container mx-auto flex flex-wrap items-center' },
                React.createElement("div", { className: 'flex flex-1 text-white font-extrabold' },
                    React.createElement(Link, { to: "/", className: 'font-light leading-tight text-2xl' }, "Broadcast Box"))))),
        React.createElement("main", { className: `${navbarEnabled && "pt-12 md:pt-12"}` },
            React.createElement(Outlet, null)),
        React.createElement("footer", { className: "mx-auto px-2 container py-6" },
            React.createElement("ul", { className: "flex items-center justify-center mt-3 text-sm:mt-0 space-x-4" },
                React.createElement("li", null,
                    React.createElement("a", { href: "https://github.com/Glimesh/broadcast-box", className: "hover:underline" }, "GitHub")),
                React.createElement("li", null,
                    React.createElement("a", { href: "https://pion.ly", className: "hover:underline" }, "Pion")),
                React.createElement("li", null,
                    React.createElement("a", { href: "https://glimesh.tv", className: "hover:underline" }, "Glimesh"))))));
};
export default RootWrapper;
