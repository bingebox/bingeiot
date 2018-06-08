#ifndef __FX_API_H__
#define __FX_API_H__

#include <stdint.h>

typedef struct rs_rect {
	int left;
	int top;
	int width;
	int height;
}__attribute__((aligned(4))) RS_RECT ;


typedef struct rs_point {
	float x;
	float y;
}__attribute__((aligned(4))) RS_POINT;


typedef enum rs_pixel_fmt{
	PIX_FORMAT_GRAY,
	PIX_FORMAT_BGR888,
	PIX_FORMAT_NV21,
	PIX_FORMAT_BGRA8888,
	PIX_FORMAT_I422,
	PIX_FORMAT_I420,
	PIX_FORMAT_YUV444,
}RS_PIXEL_FMT;


typedef struct rs_img {
	int width;
	int height;
	int stride;
	int imgSize;
	RS_PIXEL_FMT fmt;
	int placeholder;
	unsigned char* img;
}__attribute__((aligned(16))) RS_IMG;


typedef struct rs_face {
	RS_RECT rect;
	RS_POINT landmarks21[21];
	float yaw;
	float pitch;
	float roll;
	int trackId;
	int frameId;
	int feature_version;
	float faceFeature[512];
	int hasFeature;
	float faceScore;
	float imgScore;
	int age;
	int gender;
}__attribute__((aligned(4))) RS_FACE;


typedef struct rs_faceDet_info {
    RS_FACE face;
    RS_IMG img;
}__attribute__((aligned(4))) RS_FACEDET_INFO;


typedef enum FX_RUNMODE
{
    FX_RUNMODE_NONE    = 0,
	FX_RUNMODE_LIVEVID = 1,
    FX_RUNMODE_FACECAP = 2,   
    FX_RUNMODE_ENROLL  = 3,
	FX_RUNMODE_TRACK   = 4,
    FX_RUNMODE_UNKNOWN = 0x7F
}FX_RUNMODE;


typedef struct rs_ver {
    int major;
    int minor;
    int revision;
}__attribute__((aligned(4))) RS_VER;


typedef struct rs_idenfiResult
{
	int person_id[5];
	int face_id[5];
	float confidence[5];
	int count;
}__attribute__((aligned(4)))RS_IdenfiResult;


#ifdef __cplusplus
extern "C"{
#endif


/********************************************
 general purpose functions 
 *******************************************/
int fxInit();
#ifdef ANDROID_PLATFORM
int fxInit_fd(int fd);
#endif
int fxInit_fd(int fd);
int fxUninit();
int fxPing();
int fxVersion(RS_VER* version);

/* set/switch device running mode */
int fxSetRunMode(FX_RUNMODE mode);
/* get/switch device running mode */
FX_RUNMODE fxGetRunMode();

/* update model */
int fxUpdateModel(char* filepath);
/* update firmware */
int fxUpdateFirmware(char* filepath);


/********************************************
 YK_RUNMODE_ENROLL mode functions 
 *******************************************/
int fxExtractFeature(RS_IMG img, float *pFaceFeature, int *feature_version);
int fxCreateFace(float *pFaceFeature);
int fxCreatePerson(int face_id);
int fxPersonAddFace(int person_id, int face_id);
int fxPersonDelFace(int person_id, int face_id);
int fxPersonIdentification(float *pFaceFeature, RS_IdenfiResult *pResult);
int fxPersonDelete(int person_id);
int fxGetAlbumSize();
int fxResetAlbum();
int fxImportAlbum(char *filename);
int fxFaceDetect(RS_IMG srcImg, RS_FACEDET_INFO **pFaceDetInfo, int *face_count);
int fxSetROI(RS_RECT *pRoiRect);
/* capture a frame and return frame data */
int fxCaptureFrame(RS_IMG *img);
float fxFaceVerification(float *pFaceFeature1, float *pFaceFeature2);


/********************************************
 YK_RUNMODE_TRACK mode functions 
 *******************************************/
typedef enum FX_TRACKMODE
{
	FX_RECOGNITION_HAS = 0,
	FX_RECOGNITION_NO
}FX_TRACKMODE;

int fxInitTrackMode(FX_TRACKMODE trackMode, int interval);

int fxFaceTrack(RS_FACE **pFacesArray, RS_IMG **pImgsArry, int *faceCount, int *frameID);



/********************************************
 LIVEVID mode functions 
 *******************************************/
typedef void (*FX_FRAME_READY_CALLBACK) (unsigned char *data, int length);
int fxSetH264FrameReadyCallback(FX_FRAME_READY_CALLBACK callback);


//---------------------------------
typedef struct fx_obj_info {
    RS_RECT rect;
    uint64_t timestamp;
    RS_IMG img;
}__attribute__((aligned(4))) FX_OBJ_INFO;

typedef void (*FX_OBJ_DETECTED_CALLBACK) (int pObjCount, FX_OBJ_INFO *pFaceArray, FX_OBJ_INFO *pBodyArray);
int fxSetOBjDetectedCallback(FX_OBJ_DETECTED_CALLBACK callback);


/********************************************
 TRACE mode functions 
 *******************************************/
typedef struct fx_track_info {
	RS_FACE face;
    RS_IMG img;
	int trackId;
}__attribute__((aligned(4))) FX_TRACK_INFO;


typedef void (*FX_TRACE_READY_CALLBACK) (int frameid, int faceCnt, FX_TRACK_INFO *pFaceArr);
int fxSetTrackCallback(FX_TRACE_READY_CALLBACK callback);


typedef void (*FX_HEARTBEAT_CALLBACK) (unsigned int heartbeat);
int fxSetHeartbeatCallback(FX_HEARTBEAT_CALLBACK callback);


int fxEmmcCheck();



#ifdef __cplusplus
}
#endif


#endif



