import React from "react";

interface Props {
  title?: string;
  onClick?: () => void;
}

export default function Button(props: Props) {
  return (
    <button
      className="bg-green-700 text-white px-4 py-2 rounded w-full"
      onClick={props.onClick}
    >
      {props.title}
    </button>
  );
}
