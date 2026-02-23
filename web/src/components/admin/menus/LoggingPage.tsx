import React, { useContext, useEffect, useState } from "react";
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
    <div className="flex flex-col p-6 w-full max-w-6xl">
      <h1 className="text-3xl font-bold mb-6">{locale.admin_page_logging.title}</h1>

      <div className="flex whitespace-pre-line overflow-y-scroll max-h-250">
        {response}
      </div>
    </div>
  );
}
export default LoggingPage;
