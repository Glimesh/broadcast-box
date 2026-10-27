import { useContext, useRef, type ReactNode } from "react";
import { LocaleContext } from "../../providers/LocaleProvider";

interface Props<T extends string | number> {
  title: string;
  message: string;
  placeholder?: string;
  isSecret?: boolean;
  children?: ReactNode;
  onAccept?: (result: T) => void;
  onChange?: (result: T) => void;
  initialValue?: T;

  canCloseOnBackgroundClick?: boolean;

  buttonAcceptText?: string;
}

export default function TextInputDialog<T extends string | number>(
  props: Props<T>,
) {
  const { locale } = useContext(LocaleContext);
  const valueRef = useRef<HTMLInputElement>(null);

  return (
    <div className="window w-full max-w-md" onClick={(e) => e.stopPropagation()}>
      <div className="window-bar">
        <span className="window-dots" aria-hidden="true">
          <span className="window-dot" />
          <span className="window-dot" />
          <span className="window-dot" />
        </span>
        <span>{props.title}</span>
        <span className="window-bar__spacer" />
        <span className="window-bar__code">auth ♡</span>
      </div>

      <div className="modal-body">
        <p className="modal-copy">{props.message}</p>

        <input
          className="pixel-input"
          type={props.isSecret ? "password" : "text"}
          ref={valueRef}
          placeholder={props.placeholder}
          autoFocus
          onKeyUp={(e) => {
            if (e.key === "Enter") {
              props.onAccept?.(valueRef.current?.value as T);
            }
          }}
        />

        <div className="modal-actions">
          <button
            type="button"
            className="pixel-btn pixel-btn--lime"
            onClick={() => props.onAccept?.(valueRef.current?.value as T)}
          >
            {props.buttonAcceptText !== undefined
              ? props.buttonAcceptText
              : locale.shared_component_text_input_dialog.button_accept}
          </button>
        </div>
      </div>
    </div>
  );
}
