import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
const AvailableStreams = () => {
    const apiPath = import.meta.env.VITE_API_PATH;
    const navigate = useNavigate();
    const [streams, setStreams] = useState(undefined);
    useEffect(() => {
        updateStreams();
        const interval = setInterval(() => {
            updateStreams();
        }, 5000);
        return () => clearInterval(interval);
    }, []);
    const updateStreams = () => {
        fetch(`${apiPath}/status`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        })
            .then(result => {
            if (!result.ok) {
                throw new Error('Unknown error when calling status');
            }
            if (result.status === 503) {
                throw new Error('Status API disabled');
            }
            return result.json();
        })
            .then(result => {
            setStreams(result.map((resultEntry) => ({
                streamKey: resultEntry.streamKey,
            })));
        })
            .catch(() => {
            setStreams(undefined);
        });
    };
    const onWatchStreamClick = (key) => {
        if (key !== '') {
            navigate(`/${key}`);
        }
    };
    if (streams === undefined) {
        return React.createElement(React.Fragment, null);
    }
    return (React.createElement("div", { className: "flex flex-col" },
        React.createElement("h2", { className: "font-light leading-tight text-4xl mb-2 mt-6" }, "Current Streams"),
        streams.length === 0 && React.createElement("p", { className: 'flex justify-center mt-6' }, "No streams currently available"),
        streams.length !== 0 && React.createElement("p", null, "Click a stream to join it"),
        React.createElement("div", { className: "m-2" }),
        React.createElement("div", { className: 'flex flex-col' }, streams.map((e, i) => (React.createElement("button", { key: i + '_' + e.streamKey, className: `mt-2 py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75`, onClick: () => onWatchStreamClick(e.streamKey) }, e.streamKey))))));
};
export default AvailableStreams;
