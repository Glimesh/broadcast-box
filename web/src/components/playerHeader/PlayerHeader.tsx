import React from "react";

interface PlayerHeaderProps {
  children: React.ReactNode;
  headerType: "Error" | "Warning" | "Success";
}

const PlayerHeader = (props: PlayerHeaderProps) => {
  return (
    <p className={`player-header player-header--${props.headerType.toLowerCase()}`}>
      {props.children}
    </p>
  );
};

export default PlayerHeader;
