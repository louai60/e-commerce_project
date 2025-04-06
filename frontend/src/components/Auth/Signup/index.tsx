"use client";
import React, { useState, useEffect, useReducer, useMemo, useCallback } from "react";

import Breadcrumb from "@/components/Common/Breadcrumb";
import Link from "next/link";
import { useDispatch, useSelector } from 'react-redux';
import { login, register, clearError, setError } from '@/redux/features/auth/authSlice';
import { useRouter } from 'next/navigation';
import type { RootState } from '@/redux/store';
import { AppDispatch } from '@/redux/store';
import zxcvbn from 'zxcvbn';
import { Eye, EyeOff, Check, X } from 'lucide-react';

const formReducer = (state, action) => {
  switch (action.type) {
    case 'SET_FIELD':
      return { ...state, [action.field]: action.value };
    default:
      return state;
  }
};

const Signup = () => {
  const dispatch = useDispatch<AppDispatch>();
  const router = useRouter();
  const { loading: reduxLoading, error } = useSelector((state: RootState) => state.auth); // Rename redux loading state
  
  const [formData, dispatchForm] = useReducer(formReducer, {
    email: '',
    password: '',
    confirmPassword: '',
    firstName: '',
    lastName: '',
    username: '',
    phoneNumber: ''
  });

  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [isLoading, setIsLoading] = useState(false); // Add local loading state
  const [passwordStrength, setPasswordStrength] = useState(0);
  const [validations, setValidations] = useState({
    email: true,
    password: true,
    confirmPassword: true,
    firstName: true,
    lastName: true,
    username: true,
    phoneNumber: true
  });

  useEffect(() => {
    if (formData.password) {
      const result = zxcvbn(formData.password);
      setPasswordStrength(result.score);
    }
  }, [formData.password]);

  const validateField = useCallback((name: string, value: string) => {
    let isValid = true;
    switch (name) {
      case 'email':
        isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
        break;
      case 'password':
        isValid = value.length >= 8;
        break;
      case 'confirmPassword':
        isValid = value === formData.password;
        break;
      case 'phoneNumber':
        isValid = /^\+?[1-9]\d{1,14}$/.test(value);
        break;
      default:
        isValid = value.length > 0;
    }
    setValidations(prev => ({ ...prev, [name]: isValid }));
    return isValid;
  }, [formData.password]);

  const handleBlur = useCallback((e: React.FocusEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setTouched(prev => ({ ...prev, [name]: true }));
    validateField(name, value);
  }, [validateField]);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    if (error) dispatch(clearError());
    const { name, value } = e.target;
    dispatchForm({ type: 'SET_FIELD', field: name, value });
    
    if (touched[name]) {
      validateField(name, value);
    }
    
    // Special case for confirm password validation
    if (name === 'password' && touched.confirmPassword) {
      validateField('confirmPassword', formData.confirmPassword);
    }
  }, [error, touched, validateField, formData.confirmPassword, dispatch]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate all fields before submission
    let isFormValid = true;
    const newTouched = {} as Record<string, boolean>;
    
    Object.keys(formData).forEach(key => {
      newTouched[key] = true;
      const fieldValid = validateField(key, formData[key as keyof typeof formData]);
      if (!fieldValid) isFormValid = false;
    });
    
    setTouched(newTouched);
    
    if (!isFormValid) {
      dispatch(setError('Please fix the errors in the form'));
      return;
    }

    try {
      setIsLoading(true); // Use local loading state setter
      
      const registerData = {
        Email: formData.email,
        Username: formData.username,
        Password: formData.password,
        FirstName: formData.firstName,
        LastName: formData.lastName,
        PhoneNumber: formData.phoneNumber || ""
      };

      const result = await dispatch(register(registerData)).unwrap();
      if (result) {
        router.push('/');
      }
    } catch (error: any) {
      console.error('Registration error:', error);
      dispatch(setError(error.message || 'Registration failed'));
    } finally {
      setIsLoading(false); // Use local loading state setter
    }
  }, [dispatch, formData, router, validateField]);

  const getPasswordStrengthColor = useMemo(() => {
    switch (passwordStrength) {
      case 0: return 'bg-red-500';
      case 1: return 'bg-orange-500';
      case 2: return 'bg-yellow-500';
      case 3: return 'bg-blue-500';
      case 4: return 'bg-green-500';
      default: return 'bg-gray-300';
    }
  }, [passwordStrength]);

  const getPasswordStrengthText = useMemo(() => {
    switch (passwordStrength) {
      case 0: return 'Very Weak';
      case 1: return 'Weak';
      case 2: return 'Fair';
      case 3: return 'Good';
      case 4: return 'Strong';
      default: return '';
    }
  }, [passwordStrength]);

  const renderValidationIcon = useCallback((fieldName: string) => {
    if (!touched[fieldName]) return null;
    
    return validations[fieldName as keyof typeof validations] ? (
      <Check className="absolute right-4 top-1/2 transform -translate-y-1/2 text-green-500" size={18} />
    ) : (
      <X className="absolute right-4 top-1/2 transform -translate-y-1/2 text-red-500" size={18} />
    );
  }, [touched, validations]);

  const getInputClasses = useCallback((fieldName: string) => {
    const baseClasses = "rounded-lg border bg-gray-1 placeholder:text-dark-5 w-full py-3 px-5 outline-none duration-200 focus:border-transparent focus:shadow-input focus:ring-2 focus:ring-blue/20";
    
    if (!touched[fieldName]) return `${baseClasses} border-gray-3`;
    
    return validations[fieldName as keyof typeof validations]
      ? `${baseClasses} border-green-500`
      : `${baseClasses} border-red-500`;
  }, [touched, validations]);

  return (
    <>
      <Breadcrumb title={"Signup"} pages={["Signup"]} />
      <section className="overflow-hidden py-20 bg-gray-2">
        <div className="max-w-[1170px] w-full mx-auto px-4 sm:px-8 xl:px-0">
          <div className="max-w-[570px] w-full mx-auto rounded-xl bg-white shadow-1 p-4 sm:p-7.5 xl:p-11">
            <div className="text-center mb-8">
              <h2 className="font-semibold text-xl sm:text-2xl xl:text-heading-5 text-dark mb-3">
                Create an Account
              </h2>
              <p className="text-gray-600">Join us today and get access to all features</p>
              {error && (
                <div className="mt-4 p-3 rounded-lg bg-red-50 text-red-600">
                  {error}
                </div>
              )}
            </div>

            <div className="mt-5.5">
              <form onSubmit={handleSubmit}>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                  <div className="relative">
                    <label htmlFor="firstName" className="block mb-2.5 font-medium">
                      First Name <span className="text-red-500">*</span>
                    </label>

                    <div className="relative">
                      <input
                        type="text"
                        name="firstName"
                        id="firstName"
                        value={formData.firstName}
                        onChange={handleChange}
                        onBlur={handleBlur}
                        placeholder="Enter your first name"
                        className={getInputClasses('firstName')}
                      />
                      {renderValidationIcon('firstName')}
                    </div>
                    {touched.firstName && !validations.firstName && (
                      <p className="text-red-500 text-sm mt-1">First name is required</p>
                    )}
                  </div>

                  <div className="relative">
                    <label htmlFor="lastName" className="block mb-2.5 font-medium">
                      Last Name <span className="text-red-500">*</span>
                    </label>

                    <div className="relative">
                      <input
                        type="text"
                        name="lastName"
                        id="lastName"
                        value={formData.lastName}
                        onChange={handleChange}
                        onBlur={handleBlur}
                        placeholder="Enter your last name"
                        className={getInputClasses('lastName')}
                      />
                      {renderValidationIcon('lastName')}
                    </div>
                    {touched.lastName && !validations.lastName && (
                      <p className="text-red-500 text-sm mt-1">Last name is required</p>
                    )}
                  </div>
                </div>

                <div className="mt-5">
                  <label htmlFor="username" className="block mb-2.5 font-medium">
                    Username <span className="text-red-500">*</span>
                  </label>

                  <div className="relative">
                    <input
                      type="text"
                      name="username"
                      id="username"
                      value={formData.username}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      placeholder="Choose a unique username"
                      className={getInputClasses('username')}
                    />
                    {renderValidationIcon('username')}
                  </div>
                  {touched.username && !validations.username && (
                    <p className="text-red-500 text-sm mt-1">Username is required</p>
                  )}
                </div>

                <div className="mt-5">
                  <label htmlFor="email" className="block mb-2.5 font-medium">
                    Email Address <span className="text-red-500">*</span>
                  </label>

                  <div className="relative">
                    <input
                      type="email"
                      name="email"
                      id="email"
                      value={formData.email}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      placeholder="Enter your email address"
                      className={getInputClasses('email')}
                    />
                    {renderValidationIcon('email')}
                  </div>
                  {touched.email && !validations.email && (
                    <p className="text-red-500 text-sm mt-1">Please enter a valid email address</p>
                  )}
                </div>

                <div className="mt-5">
                  <label htmlFor="phoneNumber" className="block mb-2.5 font-medium">
                    Phone Number <span className="text-red-500">*</span>
                  </label>

                  <div className="relative">
                    <input
                      type="tel"
                      name="phoneNumber"
                      id="phoneNumber"
                      value={formData.phoneNumber}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      placeholder="+1234567890"
                      className={getInputClasses('phoneNumber')}
                    />
                    {renderValidationIcon('phoneNumber')}
                  </div>
                  {touched.phoneNumber && !validations.phoneNumber && (
                    <p className="text-red-500 text-sm mt-1">Please enter a valid phone number</p>
                  )}
                </div>

                <div className="mt-5">
                  <label htmlFor="password" className="block mb-2.5 font-medium">
                    Password <span className="text-red-500">*</span>
                  </label>

                  <div className="relative">
                    <input
                      type={showPassword ? "text" : "password"}
                      name="password"
                      id="password"
                      value={formData.password}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      placeholder="Create a strong password"
                      autoComplete="new-password"
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
                  
                  {formData.password && ( 
                    <div className="mt-2">
                      <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
                        <div 
                          className={`h-full ${getPasswordStrengthColor}`}
                          style={{ width: `${(passwordStrength + 1) * 20}%` }}
                        ></div>
                      </div>
                      <p className="text-sm mt-1">
                        Password strength: <span className={`font-medium ${getPasswordStrengthColor.replace('bg-', 'text-')}`}>{getPasswordStrengthText}</span>
                      </p>
                    </div>
                  )}
                  
                  {touched.password && !validations.password && (
                    <p className="text-red-500 text-sm mt-1">Password must be at least 8 characters</p>
                  )}
                </div>

                <div className="mt-5">
                  <label htmlFor="confirmPassword" className="block mb-2.5 font-medium">
                    Confirm Password <span className="text-red-500">*</span>
                  </label>

                  <div className="relative">
                    <input
                      type={showConfirmPassword ? "text" : "password"}
                      name="confirmPassword"
                      id="confirmPassword"
                      value={formData.confirmPassword}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      placeholder="Confirm your password"
                      autoComplete="new-password"
                      className={getInputClasses('confirmPassword')}
                    />
                    <button
                      type="button"
                      onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                      className="absolute right-4 top-1/2 transform -translate-y-1/2 text-gray-500"
                    >
                      {showConfirmPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </div>
                  {touched.confirmPassword && !validations.confirmPassword && (
                    <p className="text-red-500 text-sm mt-1">Passwords do not match</p>
                  )}
                </div>

                <div className="mt-6">
                  <label className="flex items-center cursor-pointer">
                    <input type="checkbox" className="w-4 h-4 accent-blue-500" required />
                    <span className="ms-2.5 text-sm">
                      I agree to the <Link href="/terms" className="text-blue-600 hover:underline">Terms of Service</Link> and <Link href="/privacy" className="text-blue-600 hover:underline">Privacy Policy</Link>
                    </span>
                  </label>
                </div>

                <button
                  type="submit"
                  disabled={isLoading || reduxLoading} // Disable if local or redux loading is true
                  className="w-full flex justify-center font-medium text-white bg-blue-600 py-3.5 px-6 rounded-lg transition-all duration-200 hover:bg-blue-700 active:bg-blue-800 mt-7.5 disabled:opacity-50 disabled:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 relative overflow-hidden group"
                  aria-label="Create your account"
                >
                  {(isLoading || reduxLoading) ? ( // Show indicator if local or redux loading is true
                    <>
                      <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      <span className="text-white">Creating Account...</span>
                    </>
                  ) : <span className="text-white">Create Account</span>}
                </button>

                <p className="text-center mt-6 text-gray-600">
                  Already have an account?{" "}
                  <Link
                    href="/signin"
                    className="text-blue-600 font-medium ease-out duration-200 hover:underline"
                  >
                    Sign in Now
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

export default Signup;

