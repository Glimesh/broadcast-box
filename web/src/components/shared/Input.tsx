import React, { RefObject } from "react";

interface Props {
  label?: string;
  ref: RefObject<HTMLInputElement | null>;
  onKeyUp?: () => void;
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
        className={`mb-2 appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200 ${props.classNames}`}
        type='text'
        placeholder={props.placeholder}
        onKeyUp={(e => {
          if (e.key === "Enter") {
            props.onKeyUp?.()
          }
        })}
        ref={props.ref}
        autoFocus={props.hasAutofocus} />
    </div>
  );

}

