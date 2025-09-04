import React, { useState } from "react";
import { getIcon, IconType } from "./Icons";

interface Props {
  titleLeft?: string;
  titleRight?: string;
  isDisabled?: boolean;
  onClickLeft?: () => void;
  onClickRight?: () => void;

  color?: "Accept" | "Deny"
  iconLeft?: IconType
  iconRight?: IconType
  classNames?: string;
}

export default function Toggle(props: Props) {

  const [selected, setSelected] = useState<"Left" | "Right">("Left")

  return (
    <div className="flex rounded-md shadow-xs justify-center mt-6" role="group">

      <button
        type="button"
        onClick={() => {
          setSelected("Left")
          props.onClickLeft?.()
        }}
        className={`${selected === "Left" ? "bg-blue-700" : ""} flex items-center px-4 py-2 text-sm font-medium border border-gray-200 rounded-s-lg hover:text-blue-700 dark:border-gray-700 dark:text-white dark:hover:text-white dark:hover:bg-blue-700 dark:focus:ring-blue-500 dark:focus:text-white`}>

        {props.iconLeft && getIcon(props.iconLeft)}
        {props.titleLeft}
      </button>

      <button
        type="button"
        onClick={() => {
          setSelected("Right")
          props.onClickRight?.()
        }}
        className={`${selected === "Right" ? "bg-blue-700" : ""} flex items-center px-4 py-2 text-sm font-medium border border-gray-200 rounded-e-lg hover:text-blue-700 dark:border-gray-700 dark:text-white dark:hover:text-white dark:hover:bg-blue-700 dark:focus:ring-blue-500 dark:focus:text-white`}>

        {props.iconRight && getIcon(props.iconRight)}
        {props.titleRight}
      </button>

    </div>)
}
