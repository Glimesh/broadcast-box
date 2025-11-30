import React, { useContext } from "react";
import { VideoCameraSlashIcon } from "@heroicons/react/16/solid";
import { LocaleContext } from "../../../providers/LocaleProvider";

interface StatusMessageComponentProps{ 
  streamKey: string;
  state: "Loading" | "Playing" | "Offline" | "Error"
}

export const StatusMessageComponent = (props: StatusMessageComponentProps) => {
	const { locale } = useContext(LocaleContext)
  const { streamKey, state } = props

  if(state === "Playing"){
    return
  }

  return <div className="absolute w-full h-full">
					{state === "Error" && (
						<div className="relative flex z-25 w-full h-full font-light leading-tight text-4xl text-center justify-center">
							<div className='flex flex-col justify-center items-center'>
								<VideoCameraSlashIcon className="w-32 h-32" />
								{streamKey} {locale.player.message_error}
							</div>
						</div>
					)}
					{state === "Offline" && (
						<div className="relative flex z-25 w-full h-full font-light leading-tight text-4xl text-center justify-center">
							<div className='flex flex-col justify-center items-center'>
								<VideoCameraSlashIcon className="w-32 h-32" />
								{streamKey} {locale.player.message_is_not_online}
							</div>
						</div>
					)}
					{state === "Loading" && (
						<div className="relative flex z-25 w-full h-full font-light leading-tight text-4xl text-center justify-center">
							<div className='flex flex-col justify-center items-center'>
								<VideoCameraSlashIcon className="w-32 h-32" />
								{streamKey} {locale.player.message_loading_video}
							</div>
						</div>
					)}
				</div>
}
