import React from "react";

interface Props {
  title?: string;
  isDisabled?: boolean;
  onClick?: () => void;
}

export default function Button(props: Props) {
  return (
    <button
      className={`text-white px-4 py-2 rounded w-full ${props.isDisabled ? "bg-gray-600" : "hover:bg-green-500   bg-green-700"}`}
      onClick={!props.isDisabled ? props.onClick : undefined}
    >
      {props.title}
    </button>
  );
}
