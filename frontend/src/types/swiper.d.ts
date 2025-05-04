declare module 'swiper/react' {
  import React from 'react';

  // Define our own SwiperInstance interface
  interface SwiperInstance {
    slidePrev: () => void;
    slideNext: () => void;
    init: () => void;
    [key: string]: any;
  }

  interface SwiperRef {
    swiper: SwiperInstance;
  }

  interface SwiperProps {
    [key: string]: any;
  }

  export const Swiper: React.ForwardRefExoticComponent<
    SwiperProps & React.RefAttributes<SwiperRef>
  >;

  export const SwiperSlide: React.FC<React.PropsWithChildren<any>>;
}
