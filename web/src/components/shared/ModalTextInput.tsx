import React, { useContext, useEffect, useRef } from "react";
import { LocaleContext } from "../../providers/LocaleProvider";
import ErrorMessagePanel from "./ErrorMessagePanel";

interface Props<T extends string | number> {
  title: string;
  errorTitle?: string;
  message: string;
  errorMessage?: string;
  placeholder?: string;
  children?: React.ReactNode;
  isOpen: boolean;
  onAccept?: (result: T) => void;
  onDeny?: () => void;
  onChange?: (result: T) => void;
  onClose?: () => void;
  initialValue?: T;

  canCloseOnBackgroundClick?: boolean;
}

export default function ModalTextInput<T extends string | number>(
  props: Props<T>,
) {
  const { locale } = useContext(LocaleContext);
  const { onClose } = props;
  const valueRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!props.isOpen) {
      onClose?.();
    }
  }, [onClose, props.isOpen]);

  if (!props.isOpen) {
    return <></>;
  }

  return (
    <div className="modal-backdrop animate-fade-in">
      <div
        className="absolute inset-0"
        onClick={() => {
          if (props.canCloseOnBackgroundClick) {
            props.onDeny?.();
            props.onClose?.();
          }
        }}
      />

      <div
        className={`window modal-window animate-scale-in ${props.children ? "modal-window--wide" : ""}`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="window-bar">
          <span className="window-dots" aria-hidden="true">
            <span className="window-dot" />
            <span className="window-dot" />
            <span className="window-dot" />
          </span>
          <span>{props.title}</span>
          <span className="window-bar__spacer" />
          <span className="window-bar__code">input ♡</span>
        </div>

        <div className="modal-body">
          <p className="modal-copy">{props.message}</p>

          {props.errorMessage !== undefined && props.errorMessage !== "" && (
            <ErrorMessagePanel
              className="mb-3"
              title={props.errorTitle ?? "Error"}
              message={props.errorMessage}
            />
          )}

          <input
            className="pixel-input"
            type="text"
            ref={valueRef}
            defaultValue={props.initialValue}
            placeholder={props.placeholder}
            onKeyUp={(evt) => evt.key === "Enter" ? props.onAccept?.(valueRef.current?.value as T) : null}
            autoFocus
          />

          {props.children && <div className="mt-3">{props.children}</div>}

          <div className="modal-actions">
            <button
              type="button"
              className="pixel-btn pixel-btn--lime"
              onClick={() => props.onAccept?.(valueRef.current?.value as T)}
            >
              {locale.shared_component_text_input_modal.button_accept}
            </button>
            <button
              type="button"
              onClick={() => {
                props.onDeny?.();
                props.onClose?.();
              }}
              className="pixel-btn pixel-btn--ghost"
            >
              close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
