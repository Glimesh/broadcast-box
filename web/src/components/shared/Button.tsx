import { getIcon, IconType } from "./Icons";

const colorMapping = {
  Accept: "pixel-btn pixel-btn--lime",
  Deny: "pixel-btn pixel-btn--danger",
  Default: "pixel-btn",
  Disabled: "pixel-btn",
};

interface Props {
  title?: string;
  subTitle?: string;
  isDisabled?: boolean;
  onClick?: () => void;

  color?: "Accept" | "Deny";
  iconLeft?: IconType;
  iconRight?: IconType;
  classNames?: string;

  stretch?: boolean;
  center?: boolean;
}

export default function Button(props: Props) {
  const color = props.isDisabled ? colorMapping.Disabled
    : props.color === "Accept" ? colorMapping.Accept
      : props.color === "Deny" ? colorMapping.Deny
        : colorMapping.Default;

  return (
    <button
      type="button"
      className={`${color} ${props.stretch ? "pixel-btn--wide" : ""} ${props.center === true ? "justify-center" : ""} ${props.classNames ?? ""}`}
      onClick={!props.isDisabled ? props.onClick : undefined}
      disabled={props.isDisabled}
    >
      {props.iconLeft && <span className="shrink-0">{getIcon(props.iconLeft)}</span>}

      <span className="flex flex-col items-start">
        <span>{props.title}</span>
        {props.subTitle && <span className="mt-0.5 text-[0.48rem] opacity-70">{props.subTitle}</span>}
      </span>

      {props.iconRight && <span className="shrink-0">{getIcon(props.iconRight)}</span>}
    </button>
  );
}
