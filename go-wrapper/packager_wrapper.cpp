// C wrapper for the shaka-packager C++ library to enable Go bindings
// This file provides a C interface around the C++ Packager class

#include <packager/packager.h>
#include <memory>
#include <vector>
#include <string>
#include <cstring>
#include <cstdlib>

extern "C" {

// Forward declarations matching the Go file
struct PackagerWrapper {
    std::unique_ptr<shaka::Packager> packager;
};

struct PackagingParamsWrapper {
    shaka::PackagingParams params;
};

struct StreamDescriptorWrapper {
    shaka::StreamDescriptor descriptor;
};

// Packager functions
PackagerWrapper* packager_create() {
    try {
        auto wrapper = new PackagerWrapper();
        wrapper->packager = std::make_unique<shaka::Packager>();
        return wrapper;
    } catch (...) {
        return nullptr;
    }
}

void packager_destroy(PackagerWrapper* p) {
    delete p;
}

int packager_initialize(PackagerWrapper* p, PackagingParamsWrapper* params, 
                       StreamDescriptorWrapper* streams, int stream_count) {
    if (!p || !p->packager || !params || !streams || stream_count <= 0) {
        return 3; // StatusInvalidArgument
    }
    
    try {
        std::vector<shaka::StreamDescriptor> stream_descriptors;
        stream_descriptors.reserve(stream_count);
        
        for (int i = 0; i < stream_count; i++) {
            stream_descriptors.push_back(streams[i].descriptor);
        }
        
        shaka::Status status = p->packager->Initialize(params->params, stream_descriptors);
        
        if (status.ok()) {
            return 0; // StatusOK
        } else {
            // Map C++ status to our status codes
            return 1; // StatusUnknown for simplicity
        }
    } catch (...) {
        return 11; // StatusInternal
    }
}

int packager_run(PackagerWrapper* p) {
    if (!p || !p->packager) {
        return 3; // StatusInvalidArgument
    }
    
    try {
        shaka::Status status = p->packager->Run();
        return status.ok() ? 0 : 1;
    } catch (...) {
        return 11; // StatusInternal
    }
}

void packager_cancel(PackagerWrapper* p) {
    if (p && p->packager) {
        p->packager->Cancel();
    }
}

char* packager_get_library_version() {
    try {
        std::string version = shaka::Packager::GetLibraryVersion();
        char* result = (char*)malloc(version.length() + 1);
        if (result) {
            strcpy(result, version.c_str());
        }
        return result;
    } catch (...) {
        return nullptr;
    }
}

// PackagingParams functions
PackagingParamsWrapper* packaging_params_create() {
    try {
        return new PackagingParamsWrapper();
    } catch (...) {
        return nullptr;
    }
}

void packaging_params_destroy(PackagingParamsWrapper* p) {
    delete p;
}

void packaging_params_set_temp_dir(PackagingParamsWrapper* p, const char* temp_dir) {
    if (p && temp_dir) {
        p->params.temp_dir = std::string(temp_dir);
    }
}

void packaging_params_set_output_media_info(PackagingParamsWrapper* p, int output_media_info) {
    if (p) {
        p->params.output_media_info = (output_media_info != 0);
    }
}

void packaging_params_set_single_threaded(PackagingParamsWrapper* p, int single_threaded) {
    if (p) {
        p->params.single_threaded = (single_threaded != 0);
    }
}

// StreamDescriptor functions
StreamDescriptorWrapper* stream_descriptor_create() {
    try {
        return new StreamDescriptorWrapper();
    } catch (...) {
        return nullptr;
    }
}

void stream_descriptor_destroy(StreamDescriptorWrapper* s) {
    delete s;
}

void stream_descriptor_set_input(StreamDescriptorWrapper* s, const char* input) {
    if (s && input) {
        s->descriptor.input = std::string(input);
    }
}

void stream_descriptor_set_stream_selector(StreamDescriptorWrapper* s, const char* stream_selector) {
    if (s && stream_selector) {
        s->descriptor.stream_selector = std::string(stream_selector);
    }
}

void stream_descriptor_set_output(StreamDescriptorWrapper* s, const char* output) {
    if (s && output) {
        s->descriptor.output = std::string(output);
    }
}

void stream_descriptor_set_segment_template(StreamDescriptorWrapper* s, const char* segment_template) {
    if (s && segment_template) {
        s->descriptor.segment_template = std::string(segment_template);
    }
}

} // extern "C"