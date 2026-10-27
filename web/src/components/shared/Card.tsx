import React from "react";
import Button from "./Button";

interface Props {
  title?: string;
  subTitle?: string;
  placeholder?: string;
  children?: React.ReactNode;
  classNames?: string;

  onAccept?: () => void;
  onDeny?: () => void;
}

export default function Card(props: Props) {
  return (
    <div className={`window ${props.classNames ?? ""}`} onClick={(e) => e.stopPropagation()}>
      {!!props.title && (
        <div className="window-bar">
          <span className="window-dots" aria-hidden="true">
            <span className="window-dot" />
            <span className="window-dot" />
            <span className="window-dot" />
          </span>
          <span>{props.title}</span>
          <span className="window-bar__spacer" />
          <span className="window-bar__code">ok ♡</span>
        </div>
      )}

      <div className="window-body">
        {!!props.subTitle && <p className="panel-copy">{props.subTitle}</p>}
        {props.children}

        {props.onAccept !== undefined && (
          <div className="modal-actions">
            <Button title="Accept" color="Accept" onClick={() => props.onAccept?.()} />
            {props.onDeny !== undefined && <Button title="Close" onClick={() => props.onDeny?.()} />}
          </div>
        )}
      </div>
    </div>
  );
}
