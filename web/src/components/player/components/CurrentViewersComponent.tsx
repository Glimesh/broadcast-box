import React from "react";
import { UsersIcon } from "@heroicons/react/20/solid";

interface CurrentViewersComponentProps {
  currentViewersCount: number;
}

const CurrentViewersComponent = (props: CurrentViewersComponentProps) => {
  const { currentViewersCount } = props;

  return (
    <div className={"flex flex-row items-center gap-1"}>
      <UsersIcon className={"size-4"} />
      {currentViewersCount}
    </div>
  )
}

export default CurrentViewersComponent
