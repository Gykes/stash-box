import { FC, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { Button, Col, Form, Row } from "react-bootstrap";
import { yupResolver } from "@hookform/resolvers/yup";
import * as yup from "yup";
import cx from "classnames";

import { ROUTE_REGISTER, ROUTE_FORGOT_PASSWORD } from "src/constants/route";
import { getPlatformURL, getCredentialsSetting } from "src/utils/createClient";

import "./App.scss";
import { useCurrentUser } from "./hooks";

const schema = yup.object({
  username: yup.string().required("Username is required"),
  password: yup.string().required("Password is required"),
});
type LoginFormData = yup.InferType<typeof schema>;

const Messages: Record<string, string> = {
  "password-reset": "Password successfully reset",
  "account-created": "Account successfully created",
};

const Login: FC = () => {
  const [loading, setLoading] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const [loginError, setLoginError] = useState("");
  const msg = new URLSearchParams(location.search).get("msg");
  const redirect = new URLSearchParams(location.search).get("redirect");
  const { isAuthenticated } = useCurrentUser();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: yupResolver(schema),
  });

  if (isAuthenticated) navigate("/");

  const onSubmit = async (formData: LoginFormData) => {
    setLoading(true);

    const body = new FormData();
    body.append("username", formData.username);
    body.append("password", formData.password);

    const res = await fetch(`${getPlatformURL()}login`, {
      method: "POST",
      body,
      credentials: getCredentialsSetting(),
    }).finally(() => setLoading(false));

    const returnURL = decodeURIComponent(redirect ?? "") || "/";
    if (res.ok) window.location.replace(returnURL);
    else setLoginError("Access denied");
  };

  return (
    <div className="LoginPrompt">
      <Form
        className="align-self-center col-4 mx-auto"
        onSubmit={handleSubmit(onSubmit)}
      >
        <Form.Floating>
          <Form.Control
            className={cx({ "is-invalid": errors?.username })}
            placeholder="Username"
            {...register("username")}
          />
          <Form.Label>Username</Form.Label>
          <div className="invalid-feedback text-end">
            {errors?.username?.message}
          </div>
        </Form.Floating>
        <Form.Floating className="my-3">
          <Form.Control
            type="password"
            className={cx({ "is-invalid": errors?.password })}
            placeholder="Password"
            {...register("password")}
          />
          <Form.Label>Password</Form.Label>
          <div className="invalid-feedback text-end">
            {errors?.password?.message}
          </div>
        </Form.Floating>
        <Row>
          <Col xs={9}>
            <div>
              <Link to={ROUTE_REGISTER}>
                <small>Register</small>
              </Link>
            </div>
            <div>
              <Link to={ROUTE_FORGOT_PASSWORD}>
                <small>Forgot Password</small>
              </Link>
            </div>
          </Col>
          <Col xs={3} className="d-flex justify-content-end">
            <div>
              <Button type="submit" className="login-button" disabled={loading}>
                Login
              </Button>
            </div>
          </Col>
        </Row>
        <Row>
          <p className="col text-end text-danger">{loginError}</p>
        </Row>
        <Row>
          <p className="col text-end text-success">
            {Messages[msg ?? ""] ?? ""}
          </p>
        </Row>
      </Form>
    </div>
  );
};

export default Login;
