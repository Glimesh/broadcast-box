import React, { Dispatch, SetStateAction } from "react";

interface Props {
  label?: string;
  value?: string;
  setValue: Dispatch<SetStateAction<string>>;

  onKeyUp?: () => void;
  onEnterKeyUp?: () => void;
  placeholder?: string;
  classNames?: string;
  hasAutofocus?: boolean;
}

export default function Input(props: Props) {

  return (
    <div>
      {props.label && (
        <label className='block text-sm font-bold mb-2'>
          {props.label}
        </label>
      )}

      <input
        type='text'
        value={props.value}
        placeholder={props.placeholder}
        autoFocus={props.hasAutofocus}
        onChange={(e) => props.setValue(() => (e.target.value))}
        onKeyUp={(e => {
          if (e.key === "Enter") {
            props.onEnterKeyUp?.()
          } else {
            props.onKeyUp?.()
          }
        })}
        className={`mb-2 appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200 ${props.classNames}`}
      />
    </div>
  );
}

