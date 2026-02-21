import { useContext } from "react";
import { LocaleContext } from "../../../providers/LocaleProvider";

interface StreamMOTDProps {
  isOnline: boolean;
  motd: string;
  className?: string;
}
export const StreamMOTD = (props: StreamMOTDProps) =>{
	const { isOnline, motd, className = '' } = props;
	const { locale } = useContext(LocaleContext)

  return (
		<div className={`w-full ${className}`}>
			<div className="relative h-5">
					<div className={`absolute inset-0 transition-opacity duration-300 text-gray-400 ${isOnline ? "opacity-100" : "opacity-0"}`} >
						{motd}
					</div>

					<div className={`absolute inset-0 transition-opacity duration-300 text-red-400 font-semibold ${!isOnline ? "opacity-100" : "opacity-0"}`} >
						<div className='flex space-x-4'>
						{locale.player.stream_status_offline}
						</div>
					</div>
			</div>
		</div>)
}
