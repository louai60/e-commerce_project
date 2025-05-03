"use client";
import React, { useState, useReducer, useMemo, useCallback, Suspense } from "react";

import Link from "next/link";
// Remove Redux imports
// import { useDispatch, useSelector } from 'react-redux';
// import { login, clearError } from '@/redux/features/auth/authSlice';
import { useRouter, useSearchParams } from 'next/navigation';
// import type { RootState } from '@/redux/store';
// import { AppDispatch } from '@/redux/store';
import { Eye, EyeOff } from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext'; // Import useAuth
import { AuthService } from '@/services/auth.service'; // Import AuthService

const Breadcrumb = React.lazy(() => import("@/components/Common/Breadcrumb"));

const initialState = {
  formData: {
    email: '',
    password: ''
  },
  touched: {},
  validations: {
    email: true,
    password: true
  }
};

const reducer = (state, action) => {
  switch (action.type) {
    case 'SET_FORM_DATA':
      return {
        ...state,
        formData: {
          ...state.formData,
          [action.name]: action.value
        }
      };
    case 'SET_TOUCHED':
      return {
        ...state,
        touched: {
          ...state.touched,
          [action.name]: true
        }
      };
    case 'SET_VALIDATIONS':
      return {
        ...state,
        validations: {
          ...state.validations,
          [action.name]: action.isValid
        }
      };
    default:
      return state;
  }
};

const Signin = () => {
  // Remove Redux hooks
  // const dispatch = useDispatch<AppDispatch>();
  const router = useRouter();
  const searchParams = useSearchParams();
  // const { loading, error } = useSelector((state: RootState) => state.auth);
  const { login: contextLogin } = useAuth(); // Get login function from context

  // Local state for loading and error during sign-in attempt
  const [localLoading, setLocalLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  const [state, dispatchState] = useReducer(reducer, initialState);
  const [showPassword, setShowPassword] = useState(false);

  const validateField = useCallback((name: string, value: string) => {
    let isValid = true;
    switch (name) {
      case 'email':
        isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
        break;
      case 'password':
        isValid = value.length > 0;
        break;
      default:
        isValid = value.length > 0;
    }
    dispatchState({ type: 'SET_VALIDATIONS', name, isValid });
    return isValid;
  }, [dispatchState]);

  const handleBlur = useCallback((e: React.FocusEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    dispatchState({ type: 'SET_TOUCHED', name });
    validateField(name, value);
  }, [validateField]);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    // Clear local error on change
    if (localError) setLocalError(null);
    const { name, value } = e.target;
    dispatchState({ type: 'SET_FORM_DATA', name, value });

    if (state.touched[name]) {
      validateField(name, value);
    }
  }, [localError, state, validateField]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError(null); // Clear previous errors

    let isFormValid = true;
    const newTouched = {} as Record<string, boolean>;

    Object.keys(state.formData).forEach(key => {
      newTouched[key] = true; // Mark all as touched on submit attempt
      const fieldValid = validateField(key, state.formData[key as keyof typeof state.formData]);
      if (!fieldValid) isFormValid = false;
    });

    // Update touched state for all fields at once
    dispatchState({ type: 'SET_TOUCHED', name: '', value: newTouched }); // Adjust reducer if needed or set individually

    if (!isFormValid) {
      return;
    }

    setLocalLoading(true); // Start loading indicator
    try {
      // Call AuthService directly
      const response = await AuthService.login(state.formData);

      if (response.access_token && response.user) {
        // Call context login to update global state and localStorage
        contextLogin(response.access_token, response.user);

        // Redirect after successful login
        const callbackUrl = searchParams?.get('callbackUrl') || '/';
        router.push(callbackUrl);
      } else {
         // Should not happen if AuthService.login throws on failure, but good practice
         setLocalError('Login failed. Please try again.');
      }
    } catch (error: any) {
      console.error('Login error:', error);
      setLocalError(error.message || 'An unexpected error occurred.'); // Set local error state
    } finally {
      setLocalLoading(false); // Stop loading indicator
    }
  }, [contextLogin, state, router, searchParams, validateField]); // Added state dependency

  const getInputClasses = useMemo(() => (fieldName: string) => {
    const baseClasses = "rounded-lg border bg-gray-1 placeholder:text-dark-5 w-full py-3 px-5 outline-none duration-200 focus:border-transparent focus:shadow-input focus:ring-2 focus:ring-blue/20";

    if (!state.touched[fieldName]) return `${baseClasses} border-gray-3`;

    return state.validations[fieldName as keyof typeof state.validations]
      ? `${baseClasses} border-green-500`
      : `${baseClasses} border-red-500`;
  }, [state]);

  return (
    <>
      <Suspense fallback={<div>Loading...</div>}>
        <Breadcrumb title={"Signin"} pages={["Signin"]} />
      </Suspense>
      <section className="overflow-hidden py-20 bg-gray-2">
        <div className="max-w-[1170px] w-full mx-auto px-4 sm:px-8 xl:px-0">
          <div className="max-w-[570px] w-full mx-auto rounded-xl bg-white shadow-1 p-4 sm:p-7.5 xl:p-11">
            <div className="text-center mb-8">
              <h2 className="font-semibold text-xl sm:text-2xl xl:text-heading-5 text-dark mb-3">
                Sign In to Your Account
              </h2>
              <p className="text-gray-600">Welcome back! Please enter your credentials</p>
              {/* Use localError for display */}
              {localError && (
                <div className="mt-4 p-3 rounded-lg bg-red-light-5 text-red">
                  {localError}
                </div>
              )}
            </div>

            <div>
              <form onSubmit={handleSubmit}>
                <div className="mb-5">
                  <label htmlFor="email" className="block mb-2.5 font-medium">
                    Email <span className="text-red-500">*</span>
                  </label>

                  <div className="relative">
                    <input
                      type="email"
                      name="email"
                      id="email"
                      value={state.formData.email}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      placeholder="Enter your email"
                      className={getInputClasses('email')}
                    />
                  </div>
                  {state.touched.email && !state.validations.email && (
                    <p className="text-red-500 text-sm mt-1">Please enter a valid email address</p>
                  )}
                </div>

                <div className="mb-5">
                  <div className="flex justify-between mb-2.5">
                    <label htmlFor="password" className="font-medium">
                      Password <span className="text-red-500">*</span>
                    </label>
                    <Link href="/forgot-password" className="text-sm text-blue-600 hover:underline">
                      Forgot Password?
                    </Link>
                  </div>

                  <div className="relative">
                    <input
                      type={showPassword ? "text" : "password"}
                      name="password"
                      id="password"
                      value={state.formData.password}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      placeholder="Enter your password"
                      autoComplete="current-password"
                      className={getInputClasses('password')}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-4 top-1/2 transform -translate-y-1/2 text-gray-500"
                    >
                      {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </div>
                  {state.touched.password && !state.validations.password && (
                    <p className="text-red-500 text-sm mt-1">Password is required</p>
                  )}
                </div>

                <div className="mt-6">
                  <label className="flex items-center cursor-pointer">
                    <input type="checkbox" className="w-4 h-4 accent-blue-500" />
                    <span className="ms-2.5 text-sm">Remember me</span>
                  </label>
                </div>

                <button
                   type="submit"
                   disabled={localLoading} // Use localLoading
                   className="w-full flex justify-center font-medium text-blue-600 bg-blue-600 py-3.5 px-6 rounded-lg transition-all duration-200 hover:bg-blue-700 active:bg-blue-800 mt-7.5 disabled:opacity-50 disabled:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 relative overflow-hidden group"
                   aria-label="Sign in to your account"
                >
                  {/* Use localLoading for button text/spinner */}
                  {localLoading ? (
                    <>
                      <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      <span className="text-white">Signing in...</span>
                    </>
                  ) : <span className="text-white">Sign in</span>}
                </button>

                <div className="mt-8 relative flex items-center justify-center">
                  <span className="absolute left-0 top-1/2 h-px w-full bg-gray-300"></span>
                  <span className="relative bg-white px-4 text-sm text-gray-500">
                    Or continue with
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-4 mt-5">
                  <button
                    type="button"
                    className="flex items-center justify-center border border-gray-300 rounded-lg py-2.5 px-4 hover:bg-gray-50 hover:border-gray-400 active:bg-gray-100 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-2"
                    aria-label="Sign in with Google"
                  >
                    <svg className="h-5 w-5 mr-2" viewBox="0 0 24 24">
                      <path
                        fill="#EA4335"
                        d="M5.26620003,9.76452941 C6.19878754,6.93863203 8.85444915,4.90909091 12,4.90909091 C13.6909091,4.90909091 15.2181818,5.50909091 16.4181818,6.49090909 L19.9090909,3 C17.7818182,1.14545455 15.0545455,0 12,0 C7.27006974,0 3.1977497,2.69829785 1.23999023,6.65002441 L5.26620003,9.76452941 Z"
                      />
                      <path
                        fill="#34A853"
                        d="M16.0407269,18.0125889 C14.9509167,18.7163016 13.5660892,19.0909091 12,19.0909091 C8.86648613,19.0909091 6.21911939,17.076871 5.27698177,14.2678769 L1.23746264,17.3349879 C3.19279051,21.2936293 7.26500293,24 12,24 C14.9328362,24 17.7353462,22.9573905 19.834192,20.9995801 L16.0407269,18.0125889 Z"
                      />
                      <path
                        fill="#4A90E2"
                        d="M19.834192,20.9995801 C22.0291676,18.9520994 23.4545455,15.903663 23.4545455,12 C23.4545455,11.2909091 23.3454545,10.5818182 23.1272727,9.90909091 L12,9.90909091 L12,14.7272727 L18.4363636,14.7272727 C18.1187732,16.6863305 17.2662994,18.2090752 16.0407269,19.0125889 L19.834192,20.9995801 Z"
                      />
                      <path
                        fill="#FBBC05"
                        d="M5.27698177,14.2678769 C5.03832634,13.556323 4.90909091,12.7937589 4.90909091,12 C4.90909091,11.2182781 5.03443647,10.4668121 5.26620003,9.76452941 L1.23999023,6.65002441 C0.43658717,8.26043162 0,10.0753848 0,12 C0,13.9195484 0.444780743,15.7301709 1.23746264,17.3349879 L5.27698177,14.2678769 Z"
                      />
                    </svg>
                    <span className="font-medium">Google</span>
                  </button>

                  <button
                    type="button"
                    className="flex items-center justify-center border border-gray-300 rounded-lg py-2.5 px-4 hover:bg-gray-50 hover:border-gray-400 active:bg-gray-100 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-2"
                    aria-label="Sign in with Facebook"
                  >
                    <svg className="h-5 w-5 mr-2 text-[#1877F2]" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M13.397,20.997v-8.196h2.765l0.411-3.209h-3.176V7.548c0-0.926,0.258-1.56,1.587-1.56h1.684V3.127 C15.849,3.039,15.025,2.997,14.201,3c-2.444,0-4.122,1.492-4.122,4.231v2.355H7.332v3.209h2.753v8.202H13.397z" />
                    </svg>
                    <span className="font-medium">Facebook</span>
                  </button>
                </div>

                <p className="text-center mt-6 text-gray-600">
                  Don't have an account?{" "}
                  <Link
                    href="/signup"
                    className="text-blue-600 font-medium ease-out duration-200 hover:underline"
                  >
                    Create an account
                  </Link>
                </p>
              </form>
            </div>
          </div>
        </div>
      </section>
    </>
  );
};

export default Signin;