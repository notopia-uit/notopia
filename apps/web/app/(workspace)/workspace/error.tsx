'use client'
import ErrorBlock from '@notopia-uit/ui/components/error-button';

const ErrorPage = () => {
  return (
    <div
      className="
      grid min-h-screen w-full
      xl:grid-cols-2
    "
    >
      <ErrorBlock />
      <div
        className="
        relative hidden
        xl:block
      "
      >
        <img
          src="https://ui.shadcn.com/placeholder.svg"
          alt="placeholder image"
          className="
            absolute inset-0 h-full object-cover
            dark:brightness-[0.95] dark:invert
          "
        />
      </div>
    </div>
  );
};

export default ErrorPage;
