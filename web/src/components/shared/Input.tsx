import { Dispatch, SetStateAction } from "react";

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
    <div className={props.classNames}>
      {props.label && <label className="pixel-label">{props.label}</label>}

      <input
        type="text"
        value={props.value}
        placeholder={props.placeholder}
        autoFocus={props.hasAutofocus}
        onChange={(e) => props.setValue(() => e.target.value)}
        onKeyUp={(e) => {
          if (e.key === "Enter") {
            props.onEnterKeyUp?.();
          } else {
            props.onKeyUp?.();
          }
        }}
        className="pixel-input"
      />
    </div>
  );
}
