import { UsersIcon } from "@heroicons/react/20/solid";

interface CurrentViewersComponentProps {
  currentViewersCount: number;
}

const CurrentViewersComponent = (props: CurrentViewersComponentProps) => {
  const { currentViewersCount } = props;

  return (
    <div className="current-viewers flex flex-row items-center gap-1.5 text-sm">
      <UsersIcon className="w-4 h-4" />
      <span className="text-[0.85rem] tracking-wider">{currentViewersCount}</span>
    </div>
  );
};

export default CurrentViewersComponent;
