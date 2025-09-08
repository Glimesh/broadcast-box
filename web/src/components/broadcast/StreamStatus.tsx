import React, { useContext } from "react";
import Card from "../shared/Card";
import { UsersIcon } from "@heroicons/react/20/solid";
import { LocaleContext } from "../../providers/LocaleProvider";

interface StreamStatusProps {
  currentViewerCount: number
}

export default function StreamStatus(props: StreamStatusProps) {
  const { locale } = useContext(LocaleContext)

  return (
    <Card
      title="Stream status"
      subTitle='Current stream status'
    >
      {/* Status bar */}
      <div className={"flex flex-row items-center gap-8"}>
        <div className='font-medium'>
          {locale.stream_status.message_current_viewers}
        </div>
        <div className='flex flex-row items-center gap-1'>
          <UsersIcon className={"size-4"} />
          {props.currentViewerCount}
        </div>
      </div>
    </Card>
  )
}
