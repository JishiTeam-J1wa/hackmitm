import React, { createContext, useContext, useState, useCallback, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Copy,
  Send,
  Trash2,
  Repeat,
  Crosshair,
  Flag,
  Eye,
  Download,
  Upload,
  ChevronRight,
  AlertTriangle,
  CheckCircle,
  Shield,
  Zap,
  Target,
  FileText,
  Link,
  Code
} from 'lucide-react';
import { clsx } from 'clsx';

// Types
export interface MenuItem {
  id: string;
  label: string;
  icon?: React.ReactNode;
  shortcut?: string;
  disabled?: boolean;
  danger?: boolean;
  divider?: boolean;
  children?: MenuItem[];
  action?: () => void | Promise<void>;
}

interface ContextMenuContextValue {
  show: (x: number, y: number, items: MenuItem[], target?: HTMLElement) => void;
  hide: () => void;
}

// Context
const ContextMenuContext = createContext<ContextMenuContextValue | null>(null);

// Hook
export const useContextMenu = () => {
  const context = useContext(ContextMenuContext);
  if (!context) {
    throw new Error('useContextMenu must be used within ContextMenuProvider');
  }
  return context;
};

// Provider
export const ContextMenuProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, setState] = useState<{
    visible: boolean;
    x: number;
    y: number;
    items: MenuItem[];
    target?: HTMLElement;
  }>({
    visible: false,
    x: 0,
    y: 0,
    items: [],
  });

  const show = useCallback((x: number, y: number, items: MenuItem[], target?: HTMLElement) => {
    setState({ visible: true, x, y, items, target });
  }, []);

  const hide = useCallback(() => {
    setState((prev) => ({ ...prev, visible: false }));
  }, []);

  return (
    <ContextMenuContext.Provider value={{ show, hide }}>
      {children}
      <ContextMenuPortal {...state} onHide={hide} />
    </ContextMenuContext.Provider>
  );
};

// Portal for rendering context menu
interface ContextMenuPortalProps {
  visible: boolean;
  x: number;
  y: number;
  items: MenuItem[];
  onHide: () => void;
}

const ContextMenuPortal: React.FC<ContextMenuPortalProps> = ({ visible, x, y, items, onHide }) => {
  const menuRef = useRef<HTMLDivElement>(null);
  const [position, setPosition] = useState({ x, y });
  const [submenuOpen, setSubmenuOpen] = useState<string | null>(null);

  // Calculate position to prevent overflow
  useEffect(() => {
    if (visible && menuRef.current) {
      const menu = menuRef.current;
      const rect = menu.getBoundingClientRect();
      const viewportWidth = window.innerWidth;
      const viewportHeight = window.innerHeight;

      let newX = x;
      let newY = y;

      if (x + rect.width > viewportWidth - 10) {
        newX = viewportWidth - rect.width - 10;
      }
      if (y + rect.height > viewportHeight - 10) {
        newY = viewportHeight - rect.height - 10;
      }

      setPosition({ x: newX, y: newY });
    }
  }, [visible, x, y]);

  // Handle click outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onHide();
      }
    };

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onHide();
      }
    };

    if (visible) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleEscape);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [visible, onHide]);

  if (!visible) return null;

  return createPortal(
    <AnimatePresence>
      <motion.div
        ref={menuRef}
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        transition={{ duration: 0.1 }}
        style={{
          position: 'fixed',
          left: position.x,
          top: position.y,
          zIndex: 9999,
        }}
        className="min-w-[200px] max-w-[320px] bg-white rounded-lg shadow-xl border border-gray-200 py-1 overflow-hidden"
      >
        {items.map((item, index) => (
          <React.Fragment key={item.id}>
            {item.divider && index > 0 && <div className="h-px bg-gray-100 my-1" />}
            {item.children ? (
              <SubmenuItem
                item={item}
                isOpen={submenuOpen === item.id}
                onOpen={() => setSubmenuOpen(item.id)}
                onClose={() => setSubmenuOpen(null)}
              />
            ) : (
              <MenuItemComponent item={item} onClick={onHide} />
            )}
          </React.Fragment>
        ))}
      </motion.div>
    </AnimatePresence>,
    document.body
  );
};

// Menu Item Component
interface MenuItemComponentProps {
  item: MenuItem;
  onClick: () => void;
}

const MenuItemComponent: React.FC<MenuItemComponentProps> = ({ item, onClick }) => {
  const [isLoading, setIsLoading] = useState(false);

  const handleClick = async () => {
    if (item.disabled || isLoading) return;

    if (item.action) {
      setIsLoading(true);
      try {
        await item.action();
      } catch (error) {
        console.error('Menu action error:', error);
      } finally {
        setIsLoading(false);
      }
    }
    onClick();
  };

  return (
    <button
      onClick={handleClick}
      disabled={item.disabled || isLoading}
      className={clsx(
        'w-full flex items-center gap-3 px-3 py-2 text-sm transition-colors text-left',
        item.disabled
          ? 'text-gray-400 cursor-not-allowed'
          : item.danger
          ? 'text-red-600 hover:bg-red-50'
          : 'text-gray-700 hover:bg-blue-50 hover:text-blue-700'
      )}
    >
      {item.icon && (
        <span className={clsx(
          'flex-shrink-0 w-4 h-4',
          item.disabled ? 'text-gray-400' : item.danger ? 'text-red-500' : 'text-gray-500'
        )}>
          {item.icon}
        </span>
      )}
      <span className="flex-1 truncate">{item.label}</span>
      {item.shortcut && (
        <span className="text-xs text-gray-400 ml-auto pl-4">{item.shortcut}</span>
      )}
    </button>
  );
};

// Submenu Item Component
interface SubmenuItemProps {
  item: MenuItem;
  isOpen: boolean;
  onOpen: () => void;
  onClose: () => void;
}

const SubmenuItem: React.FC<SubmenuItemProps> = ({ item, isOpen, onOpen, onClose }) => {
  const submenuRef = useRef<HTMLDivElement>(null);
  const [submenuPosition, setSubmenuPosition] = useState<'right' | 'left'>('right');

  const handleMouseEnter = () => {
    if (submenuRef.current) {
      const rect = submenuRef.current.getBoundingClientRect();
      const viewportWidth = window.innerWidth;
      setSubmenuPosition(rect.right + 200 > viewportWidth ? 'left' : 'right');
    }
    onOpen();
  };

  return (
    <div
      ref={submenuRef}
      className="relative"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={onClose}
    >
      <button
        className="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-blue-50 hover:text-blue-700 transition-colors"
      >
        {item.icon && <span className="flex-shrink-0 w-4 h-4 text-gray-500">{item.icon}</span>}
        <span className="flex-1 truncate">{item.label}</span>
        <ChevronRight className="w-4 h-4 text-gray-400 ml-auto" />
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -10 }}
            transition={{ duration: 0.1 }}
            style={{
              position: 'absolute',
              top: 0,
              [submenuPosition === 'right' ? 'left' : 'right']: '100%',
            }}
            className="min-w-[200px] bg-white rounded-lg shadow-xl border border-gray-200 py-1"
          >
            {item.children?.map((child) => (
              <MenuItemComponent key={child.id} item={child} onClick={onClose} />
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

// Right-click wrapper hook
export const useRightClick = (items: MenuItem[]) => {
  const { show } = useContextMenu();

  const handleContextMenu = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      show(e.clientX, e.clientY, items, e.target as HTMLElement);
    },
    [items, show]
  );

  return { onContextMenu: handleContextMenu };
};

// Preset menu items for common actions
export const createTrafficMenuItems = (
  _traffic: any,
  actions: {
    sendToRepeater?: () => void;
    sendToIntruder?: () => void;
    sendToScanner?: () => void;
    copyUrl?: () => void;
    copyAsCurl?: () => void;
    copyRequest?: () => void;
    copyResponse?: () => void;
    saveToHistory?: () => void;
    delete?: () => void;
    repeatRequest?: () => void;
    scanForVulns?: () => void;
    addToScope?: () => void;
    removeFromScope?: () => void;
    tag?: () => void;
    export?: () => void;
  }
): MenuItem[] => {
  const items: MenuItem[] = [];

  // Send to menu
  items.push({
    id: 'send-to',
    label: '发送到',
    icon: <Send className="w-4 h-4" />,
    children: [
      {
        id: 'send-repeater',
        label: 'Repeater',
        icon: <Repeat className="w-4 h-4" />,
        shortcut: 'Ctrl+R',
        action: actions.sendToRepeater,
      },
      {
        id: 'send-intruder',
        label: 'Intruder',
        icon: <Crosshair className="w-4 h-4" />,
        shortcut: 'Ctrl+I',
        action: actions.sendToIntruder,
      },
      {
        id: 'send-scanner',
        label: 'Scanner',
        icon: <Shield className="w-4 h-4" />,
        shortcut: 'Ctrl+S',
        action: actions.sendToScanner,
      },
      {
        id: 'divider-1',
        label: '',
        divider: true,
      },
      {
        id: 'send-target',
        label: 'Target',
        icon: <Target className="w-4 h-4" />,
        action: actions.addToScope,
      },
    ],
  });

  // Copy menu
  items.push({
    id: 'copy',
    label: '复制',
    icon: <Copy className="w-4 h-4" />,
    children: [
      {
        id: 'copy-url',
        label: 'URL',
        icon: <Link className="w-4 h-4" />,
        shortcut: 'Ctrl+U',
        action: actions.copyUrl,
      },
      {
        id: 'copy-curl',
        label: 'cURL 命令',
        icon: <Code className="w-4 h-4" />,
        shortcut: 'Ctrl+Shift+C',
        action: actions.copyAsCurl,
      },
      {
        id: 'divider-2',
        label: '',
        divider: true,
      },
      {
        id: 'copy-request',
        label: '请求',
        icon: <Upload className="w-4 h-4" />,
        action: actions.copyRequest,
      },
      {
        id: 'copy-response',
        label: '响应',
        icon: <Download className="w-4 h-4" />,
        action: actions.copyResponse,
      },
    ],
  });

  items.push({ id: 'divider-3', label: '', divider: true });

  // Actions
  items.push({
    id: 'repeat',
    label: '重放请求',
    icon: <Repeat className="w-4 h-4" />,
    shortcut: 'Ctrl+Enter',
    action: actions.repeatRequest,
  });

  items.push({
    id: 'scan',
    label: '扫描漏洞',
    icon: <Zap className="w-4 h-4" />,
    shortcut: 'Ctrl+Shift+S',
    action: actions.scanForVulns,
  });

  items.push({
    id: 'tag',
    label: '添加标签',
    icon: <Flag className="w-4 h-4" />,
    action: actions.tag,
  });

  items.push({ id: 'divider-4', label: '', divider: true });

  items.push({
    id: 'export',
    label: '导出',
    icon: <Download className="w-4 h-4" />,
    action: actions.export,
  });

  items.push({
    id: 'delete',
    label: '删除',
    icon: <Trash2 className="w-4 h-4" />,
    shortcut: 'Del',
    danger: true,
    action: actions.delete,
  });

  return items;
};

// Vulnerability menu items
export const createVulnMenuItems = (
  _vuln: any,
  actions: {
    markConfirmed?: () => void;
    markFalsePositive?: () => void;
    markFixed?: () => void;
    copyDetails?: () => void;
    generateReport?: () => void;
    viewRequest?: () => void;
    delete?: () => void;
  }
): MenuItem[] => {
  return [
    {
      id: 'mark-confirmed',
      label: '标记为已确认',
      icon: <CheckCircle className="w-4 h-4 text-green-500" />,
      action: actions.markConfirmed,
    },
    {
      id: 'mark-false-positive',
      label: '标记为误报',
      icon: <AlertTriangle className="w-4 h-4 text-yellow-500" />,
      action: actions.markFalsePositive,
    },
    {
      id: 'mark-fixed',
      label: '标记为已修复',
      icon: <CheckCircle className="w-4 h-4 text-blue-500" />,
      action: actions.markFixed,
    },
    { id: 'divider-1', label: '', divider: true },
    {
      id: 'view-request',
      label: '查看请求',
      icon: <Eye className="w-4 h-4" />,
      action: actions.viewRequest,
    },
    {
      id: 'copy-details',
      label: '复制详情',
      icon: <Copy className="w-4 h-4" />,
      action: actions.copyDetails,
    },
    { id: 'divider-2', label: '', divider: true },
    {
      id: 'report',
      label: '生成报告',
      icon: <FileText className="w-4 h-4" />,
      action: actions.generateReport,
    },
    {
      id: 'delete',
      label: '删除',
      icon: <Trash2 className="w-4 h-4" />,
      danger: true,
      action: actions.delete,
    },
  ];
};

export default ContextMenuProvider;
