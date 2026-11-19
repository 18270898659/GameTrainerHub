export namespace config {
	
	export class AppConfig {
	    config_version: number;
	    base_dir: string;
	    data_dir: string;
	    download_dir: string;
	    log_dir: string;
	    max_concurrent_downloads: number;
	    auto_sync_interval_hours: number;
	    auto_sync_enabled: boolean;
	    proxy_url: string;
	    theme: string;
	    language: string;
	    log_level: string;
	    safe_launch_prompt: boolean;
	    close_to_tray: boolean;
	    remember_window_size: boolean;
	    window_width: number;
	    window_height: number;
	    remember_window_position: boolean;
	    window_x: number;
	    window_y: number;
	    notify_on_download_complete: boolean;
	    notify_on_download_failed: boolean;
	    notify_on_trainer_update: boolean;
	    global_hotkeys_enabled: boolean;
	    hotkey_toggle_window: string;
	    hotkey_open_downloads: string;
	    auto_backup_config: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config_version = source["config_version"];
	        this.base_dir = source["base_dir"];
	        this.data_dir = source["data_dir"];
	        this.download_dir = source["download_dir"];
	        this.log_dir = source["log_dir"];
	        this.max_concurrent_downloads = source["max_concurrent_downloads"];
	        this.auto_sync_interval_hours = source["auto_sync_interval_hours"];
	        this.auto_sync_enabled = source["auto_sync_enabled"];
	        this.proxy_url = source["proxy_url"];
	        this.theme = source["theme"];
	        this.language = source["language"];
	        this.log_level = source["log_level"];
	        this.safe_launch_prompt = source["safe_launch_prompt"];
	        this.close_to_tray = source["close_to_tray"];
	        this.remember_window_size = source["remember_window_size"];
	        this.window_width = source["window_width"];
	        this.window_height = source["window_height"];
	        this.remember_window_position = source["remember_window_position"];
	        this.window_x = source["window_x"];
	        this.window_y = source["window_y"];
	        this.notify_on_download_complete = source["notify_on_download_complete"];
	        this.notify_on_download_failed = source["notify_on_download_failed"];
	        this.notify_on_trainer_update = source["notify_on_trainer_update"];
	        this.global_hotkeys_enabled = source["global_hotkeys_enabled"];
	        this.hotkey_toggle_window = source["hotkey_toggle_window"];
	        this.hotkey_open_downloads = source["hotkey_open_downloads"];
	        this.auto_backup_config = source["auto_backup_config"];
	    }
	}

}

export namespace launcher {
	
	export class SecurityCardInfo {
	    trainer_id: number;
	    game_id: number;
	    version_id: number;
	    game_name: string;
	    file_name: string;
	    file_path: string;
	    file_size: number;
	    sha256: string;
	    source_name: string;
	    is_custom: boolean;
	    file_exists: boolean;
	    requires_admin: boolean;
	    security_notice: string;
	
	    static createFrom(source: any = {}) {
	        return new SecurityCardInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trainer_id = source["trainer_id"];
	        this.game_id = source["game_id"];
	        this.version_id = source["version_id"];
	        this.game_name = source["game_name"];
	        this.file_name = source["file_name"];
	        this.file_path = source["file_path"];
	        this.file_size = source["file_size"];
	        this.sha256 = source["sha256"];
	        this.source_name = source["source_name"];
	        this.is_custom = source["is_custom"];
	        this.file_exists = source["file_exists"];
	        this.requires_admin = source["requires_admin"];
	        this.security_notice = source["security_notice"];
	    }
	}

}

export namespace model {
	
	export class NetworkDiagnosis {
	    target_host: string;
	    dns_resolved: boolean;
	    ip_addresses: string;
	    http_status: number;
	    latency_ms: number;
	    is_reachable: boolean;
	    error_message?: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkDiagnosis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target_host = source["target_host"];
	        this.dns_resolved = source["dns_resolved"];
	        this.ip_addresses = source["ip_addresses"];
	        this.http_status = source["http_status"];
	        this.latency_ms = source["latency_ms"];
	        this.is_reachable = source["is_reachable"];
	        this.error_message = source["error_message"];
	    }
	}
	export class DiagnosticReport {
	    // Go type: time
	    generated_at: any;
	    app_version: string;
	    os: string;
	    arch: string;
	    download_dir: string;
	    download_dir_writable: boolean;
	    available_disk_bytes: number;
	    db_path: string;
	    db_integrity_ok: boolean;
	    network?: NetworkDiagnosis;
	    active_downloads: number;
	    local_trainers_count: number;
	    favorites_count: number;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generated_at = this.convertValues(source["generated_at"], null);
	        this.app_version = source["app_version"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.download_dir = source["download_dir"];
	        this.download_dir_writable = source["download_dir_writable"];
	        this.available_disk_bytes = source["available_disk_bytes"];
	        this.db_path = source["db_path"];
	        this.db_integrity_ok = source["db_integrity_ok"];
	        this.network = this.convertValues(source["network"], NetworkDiagnosis);
	        this.active_downloads = source["active_downloads"];
	        this.local_trainers_count = source["local_trainers_count"];
	        this.favorites_count = source["favorites_count"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DownloadHistory {
	    id: number;
	    task_id: number;
	    trainer_id: number;
	    version_id: number;
	    game_name: string;
	    trainer_name: string;
	    version_name: string;
	    file_name: string;
	    file_size: number;
	    sha256: string;
	    save_path: string;
	    file_exists: boolean;
	    // Go type: time
	    completed_at: any;
	    game_id?: number;
	    cover_url?: string;
	    detail_url?: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.task_id = source["task_id"];
	        this.trainer_id = source["trainer_id"];
	        this.version_id = source["version_id"];
	        this.game_name = source["game_name"];
	        this.trainer_name = source["trainer_name"];
	        this.version_name = source["version_name"];
	        this.file_name = source["file_name"];
	        this.file_size = source["file_size"];
	        this.sha256 = source["sha256"];
	        this.save_path = source["save_path"];
	        this.file_exists = source["file_exists"];
	        this.completed_at = this.convertValues(source["completed_at"], null);
	        this.game_id = source["game_id"];
	        this.cover_url = source["cover_url"];
	        this.detail_url = source["detail_url"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DownloadTask {
	    id: number;
	    version_id: number;
	    game_name: string;
	    trainer_name: string;
	    version_name: string;
	    file_name: string;
	    save_path: string;
	    total_bytes: number;
	    downloaded_bytes: number;
	    status: string;
	    speed_bytes_per_sec: number;
	    progress: number;
	    error_msg: string;
	    retries: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new DownloadTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.version_id = source["version_id"];
	        this.game_name = source["game_name"];
	        this.trainer_name = source["trainer_name"];
	        this.version_name = source["version_name"];
	        this.file_name = source["file_name"];
	        this.save_path = source["save_path"];
	        this.total_bytes = source["total_bytes"];
	        this.downloaded_bytes = source["downloaded_bytes"];
	        this.status = source["status"];
	        this.speed_bytes_per_sec = source["speed_bytes_per_sec"];
	        this.progress = source["progress"];
	        this.error_msg = source["error_msg"];
	        this.retries = source["retries"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TrainerVersion {
	    id: number;
	    trainer_id: number;
	    version_name: string;
	    target_game_version: string;
	    features_text: string;
	    download_url: string;
	    file_name: string;
	    file_size: number;
	    sha256: string;
	    // Go type: time
	    published_at: any;
	    is_downloaded: boolean;
	    local_path?: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new TrainerVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.trainer_id = source["trainer_id"];
	        this.version_name = source["version_name"];
	        this.target_game_version = source["target_game_version"];
	        this.features_text = source["features_text"];
	        this.download_url = source["download_url"];
	        this.file_name = source["file_name"];
	        this.file_size = source["file_size"];
	        this.sha256 = source["sha256"];
	        this.published_at = this.convertValues(source["published_at"], null);
	        this.is_downloaded = source["is_downloaded"];
	        this.local_path = source["local_path"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Trainer {
	    id: number;
	    game_id: number;
	    source_id: string;
	    external_id: string;
	    title_en: string;
	    title_cn: string;
	    cover_url?: string;
	    detail_url: string;
	    author: string;
	    latest_version: string;
	    // Go type: time
	    published_at: any;
	    // Go type: time
	    updated_at: any;
	    is_deleted: boolean;
	    versions?: TrainerVersion[];
	
	    static createFrom(source: any = {}) {
	        return new Trainer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.game_id = source["game_id"];
	        this.source_id = source["source_id"];
	        this.external_id = source["external_id"];
	        this.title_en = source["title_en"];
	        this.title_cn = source["title_cn"];
	        this.cover_url = source["cover_url"];
	        this.detail_url = source["detail_url"];
	        this.author = source["author"];
	        this.latest_version = source["latest_version"];
	        this.published_at = this.convertValues(source["published_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.is_deleted = source["is_deleted"];
	        this.versions = this.convertValues(source["versions"], TrainerVersion);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Game {
	    id: number;
	    slug: string;
	    name_en: string;
	    name_cn: string;
	    cover_url: string;
	    steam_app_id: number;
	    install_path: string;
	    aliases?: string[];
	    is_favorite: boolean;
	    detail_url?: string;
	    trainers?: Trainer[];
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Game(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.slug = source["slug"];
	        this.name_en = source["name_en"];
	        this.name_cn = source["name_cn"];
	        this.cover_url = source["cover_url"];
	        this.steam_app_id = source["steam_app_id"];
	        this.install_path = source["install_path"];
	        this.aliases = source["aliases"];
	        this.is_favorite = source["is_favorite"];
	        this.detail_url = source["detail_url"];
	        this.trainers = this.convertValues(source["trainers"], Trainer);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class IntegrityResult {
	    trainer_id: number;
	    file_path: string;
	    file_exists: boolean;
	    expected_sha256: string;
	    actual_sha256: string;
	    file_size: number;
	    is_valid: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new IntegrityResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trainer_id = source["trainer_id"];
	        this.file_path = source["file_path"];
	        this.file_exists = source["file_exists"];
	        this.expected_sha256 = source["expected_sha256"];
	        this.actual_sha256 = source["actual_sha256"];
	        this.file_size = source["file_size"];
	        this.is_valid = source["is_valid"];
	        this.message = source["message"];
	    }
	}
	export class LocalTrainer {
	    id: number;
	    game_id: number;
	    version_id?: number;
	    file_path: string;
	    file_size: number;
	    sha256: string;
	    source_name: string;
	    is_custom: boolean;
	    // Go type: time
	    last_run_at?: any;
	    run_count: number;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new LocalTrainer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.game_id = source["game_id"];
	        this.version_id = source["version_id"];
	        this.file_path = source["file_path"];
	        this.file_size = source["file_size"];
	        this.sha256 = source["sha256"];
	        this.source_name = source["source_name"];
	        this.is_custom = source["is_custom"];
	        this.last_run_at = this.convertValues(source["last_run_at"], null);
	        this.run_count = source["run_count"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LogEntry {
	    // Go type: time
	    timestamp: any;
	    level: string;
	    category: string;
	    message: string;
	    details?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.level = source["level"];
	        this.category = source["category"];
	        this.message = source["message"];
	        this.details = source["details"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SearchFilter {
	    query: string;
	    source: string;
	    local_only: boolean;
	    page: number;
	    page_size: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.source = source["source"];
	        this.local_only = source["local_only"];
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	    }
	}
	export class SearchResult {
	    game_id: number;
	    slug: string;
	    game_name_en: string;
	    game_name_cn: string;
	    cover_url: string;
	    source_id: string;
	    trainer_title: string;
	    detail_url: string;
	    version_name: string;
	    // Go type: time
	    updated_at: any;
	    is_downloaded: boolean;
	    is_online_only: boolean;
	    is_favorite: boolean;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game_id = source["game_id"];
	        this.slug = source["slug"];
	        this.game_name_en = source["game_name_en"];
	        this.game_name_cn = source["game_name_cn"];
	        this.cover_url = source["cover_url"];
	        this.source_id = source["source_id"];
	        this.trainer_title = source["trainer_title"];
	        this.detail_url = source["detail_url"];
	        this.version_name = source["version_name"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.is_downloaded = source["is_downloaded"];
	        this.is_online_only = source["is_online_only"];
	        this.is_favorite = source["is_favorite"];
	        this.score = source["score"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SyncRecord {
	    id: number;
	    source_id: string;
	    sync_type: string;
	    status: string;
	    items_added: number;
	    items_updated: number;
	    error_msg: string;
	    // Go type: time
	    started_at: any;
	    // Go type: time
	    finished_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new SyncRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source_id = source["source_id"];
	        this.sync_type = source["sync_type"];
	        this.status = source["status"];
	        this.items_added = source["items_added"];
	        this.items_updated = source["items_updated"];
	        this.error_msg = source["error_msg"];
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.finished_at = this.convertValues(source["finished_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class UpdateCheckResult {
	    has_update: boolean;
	    current_version: string;
	    latest_version: string;
	    release_notes: string;
	    download_url: string;
	    package_sha256: string;
	    published_at: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.has_update = source["has_update"];
	        this.current_version = source["current_version"];
	        this.latest_version = source["latest_version"];
	        this.release_notes = source["release_notes"];
	        this.download_url = source["download_url"];
	        this.package_sha256 = source["package_sha256"];
	        this.published_at = source["published_at"];
	    }
	}

}

export namespace trainer {
	
	export class LocalTrainerDetail {
	    local_trainer?: model.LocalTrainer;
	    game?: model.Game;
	    version?: model.TrainerVersion;
	    file_exists: boolean;
	    is_running: boolean;
	    target_game_version?: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalTrainerDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.local_trainer = this.convertValues(source["local_trainer"], model.LocalTrainer);
	        this.game = this.convertValues(source["game"], model.Game);
	        this.version = this.convertValues(source["version"], model.TrainerVersion);
	        this.file_exists = source["file_exists"];
	        this.is_running = source["is_running"];
	        this.target_game_version = source["target_game_version"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

