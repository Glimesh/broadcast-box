import React from "react";

interface ErrorHeaderProps {
  children: React.ReactNode;
}

const ErrorHeader = (props: ErrorHeaderProps) => {
  return (
    <p className={'bg-red-700 text-white text-lg text-center p-4 rounded-t-lg whitespace-pre-wrap'}>
      {props.children}
    </p>
  )
}

export default ErrorHeader;