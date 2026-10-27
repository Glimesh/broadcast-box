import { useContext, useEffect, useState } from "react";
import { LocaleContext } from "../../../providers/LocaleProvider";
import {
  clearAdminToken,
  getAdminAuthorizationHeader,
  isInvalidAdminSessionResponse,
} from "../adminAuth";

const LoggingPage = () => {
  const { locale } = useContext(LocaleContext)
  const [response, setResponse] = useState<string>()

  const getLogs = () => {
    fetch(`/api/admin/logging`, {
      method: "GET",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
      },
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken()
          return;
        }

        return result.text();
      })
      .then((result) => {
        const reversed: string[] = result?.split('\n').reverse() ?? [""]
        setResponse(() => reversed.join('\n').toString())
      });
  };

  useEffect(() => getLogs(), [])

  return (
    <div className="flex w-full flex-col gap-4">
      <h1 className="panel-heading text-base">{locale.admin_page_logging.title}</h1>

      <div className="log-view">
        {response}
      </div>
    </div>
  );
}
export default LoggingPage;
