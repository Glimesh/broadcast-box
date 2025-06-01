import React from "react";
const ErrorHeader = (props) => {
    return (React.createElement("p", { className: 'bg-red-700 text-white text-lg text-center p-4 rounded-t-lg whitespace-pre-wrap' }, props.children));
};
export default ErrorHeader;
