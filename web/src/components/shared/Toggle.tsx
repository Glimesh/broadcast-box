import React from "react";
import { getIcon, IconType } from "./Icons";

interface Props {
  titleLeft?: string;
  titleRight?: string;
  isDisabled?: boolean;
  onClickLeft?: () => void;
  onClickRight?: () => void;

  label?: string;
  selected?: "Left" | "Right"
  color?: "Accept" | "Deny"
  iconLeft?: IconType
  iconRight?: IconType
  classNames?: string;
}

export default function Toggle(props: Props) {
  return (
    <div>
      {/* Label */}
      {props.label && (
        <label className='block text-sm font-bold mb-2'>
          {props.label}
        </label>
      )}

      {/* Toggle */}
      <div className="flex rounded-md shadow-xs justify-center" role="group">
        <button
          type="button"
          onClick={() => props.onClickLeft?.()}
          className={`${props.selected === "Left" ? "bg-blue-700" : ""} flex items-center px-4 py-2 text-sm font-medium border border-gray-200 rounded-s-lg hover:text-blue-700 dark:border-gray-700 dark:text-white dark:hover:text-white dark:hover:bg-blue-700 dark:focus:ring-blue-500 dark:focus:text-white`}>

          {props.iconLeft && getIcon(props.iconLeft)}
          {props.titleLeft}
        </button>

        <button
          type="button"
          onClick={() => props.onClickRight?.()}
          className={`${props.selected === "Right" ? "bg-blue-700" : ""} flex items-center px-4 py-2 text-sm font-medium border border-gray-200 rounded-e-lg hover:text-blue-700 dark:border-gray-700 dark:text-white dark:hover:text-white dark:hover:bg-blue-700 dark:focus:ring-blue-500 dark:focus:text-white`}>

          {props.iconRight && getIcon(props.iconRight)}
          {props.titleRight}
        </button>

      </div>
    </div>)
}
