import React from "react";
import Button from "./Button";
import ErrorMessagePanel from "./ErrorMessagePanel";

interface Props {
  title: string;
  errorTitle?: string;
  message: string;
  errorMessage?: string;
  children?: React.ReactNode;
  isOpen: boolean;
  onAccept?: () => void;
  onDeny?: () => void;
  onChange?: () => void;

  canCloseOnBackgroundClick?: boolean;
}

export default function ModalMessageBox(props: Props) {
  if (!props.isOpen) {
    return <></>;
  }

  return (
    <div className="modal-backdrop animate-fade-in">
      <div
        className="absolute inset-0"
        onClick={() => props.canCloseOnBackgroundClick && props.onDeny?.()}
      />

      <div className="window modal-window animate-scale-in" onClick={(e) => e.stopPropagation()}>
        <div className="window-bar window-bar--hot">
          <span className="window-dots" aria-hidden="true">
            <span className="window-dot" />
            <span className="window-dot" />
            <span className="window-dot" />
          </span>
          <span>{props.title}</span>
          <span className="window-bar__spacer" />
          <span className="window-bar__code">confirm? ♡</span>
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

          {props.children}

          <div className="modal-actions">
            {props.onAccept !== null && (
              <Button color="Accept" title="Accept" onClick={() => props.onAccept?.()} />
            )}
            <Button title="Close" onClick={() => props.onDeny?.()} />
          </div>
        </div>
      </div>
    </div>
  );
}
