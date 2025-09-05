import React from "react";
import Card from "../shared/Card";
import { UsersIcon } from "@heroicons/react/20/solid";

interface StreamStatusProps {
  currentViewerCount: number
}

export default function StreamStatus(props: StreamStatusProps) {

  return (
    <Card
      title="Stream status"
      subTitle='Current stream status'
    >
      {/* Status bar */}
      <div className={"flex flex-row items-center gap-8"}>
        <div className='font-medium'>
          Current Viewers
        </div>
        <div className='flex flex-row items-center gap-1'>
          <UsersIcon className={"size-4"} />
          {props.currentViewerCount}
        </div>
      </div>
    </Card>
  )
}
