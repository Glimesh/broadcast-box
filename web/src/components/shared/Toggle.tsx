import { getIcon, IconType } from "./Icons";

interface Props {
  titleLeft?: string;
  titleRight?: string;
  isDisabled?: boolean;
  onClickLeft?: () => void;
  onClickRight?: () => void;

  label?: string;
  selected?: "Left" | "Right";
  color?: "Accept" | "Deny";
  iconLeft?: IconType;
  iconRight?: IconType;
  classNames?: string;
}

export default function Toggle(props: Props) {
  return (
    <div className={props.classNames}>
      {props.label && <label className="pixel-label">{props.label}</label>}

      <div className="toggle-group" role="group">
        <button
          type="button"
          disabled={props.isDisabled}
          onClick={() => props.onClickLeft?.()}
          className={`toggle-option ${props.selected === "Left" ? "is-active" : ""}`}
        >
          {props.iconLeft && <span className="shrink-0">{getIcon(props.iconLeft)}</span>}
          {props.titleLeft}
        </button>

        <button
          type="button"
          disabled={props.isDisabled}
          onClick={() => props.onClickRight?.()}
          className={`toggle-option ${props.selected === "Right" ? "is-active" : ""}`}
        >
          {props.iconRight && <span className="shrink-0">{getIcon(props.iconRight)}</span>}
          {props.titleRight}
        </button>
      </div>
    </div>
  );
}
